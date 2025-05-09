// Code adapted from go-go-labs/cmd/experiments/libgen
package libgen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog/log"
)

/* ---------- core types & constants ---------- */

// DefaultMirror is the default LibGen mirror URL
const DefaultMirror = "https://libgen.rs"

// Book represents a book from LibGen
type Book struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Year      string `json:"year"`
	Pages     string `json:"pages"`
	Extension string `json:"extension"`
	MD5       string `json:"md5"`
	FileSize  string `json:"filesize"`
}

// Paper represents a scientific paper from CrossRef/Unpaywall
type Paper struct {
	DOI     string `json:"doi"`
	Title   string `json:"title"`
	Author  string `json:"author"`
	Year    string `json:"year"`
	Journal string `json:"journal"`
	PDFLink string `json:"pdf_link"`
}

// SearchResult contains the search results and any debug logs
type SearchResult struct {
	Books  []Book
	Papers []Paper
	Logs   string
}

// DownloadResult contains the download result information and any debug logs
type DownloadResult struct {
	FilePath string
	Size     int64
	Logs     string
}

// Debug controls verbose output
var Debug bool

/* ---------- search flow ---------- */

// SearchBooks searches for books matching the query on LibGen
func SearchBooks(query string, limit int, mirror string, scientific bool) (SearchResult, error) {
	var logs bytes.Buffer
	result := SearchResult{}

	if scientific {
		return searchPapers(query, limit, &logs)
	}

	if mirror == "" {
		mirror = DefaultMirror
	}

	searchURL, ids, err := crawlIDs(query, limit, mirror)
	if err != nil {
		return result, err
	}

	jsonURL, raw, books, err := fetchDetails(ids, mirror)
	if err != nil {
		return result, err
	}

	result.Books = books

	if Debug {
		fmt.Fprintf(&logs, "# search.php URL:\n%s\n", searchURL)
		fmt.Fprintf(&logs, "# json.php URL:\n%s\n", jsonURL)
		fmt.Fprintf(&logs, "# raw JSON:\n")
		var pretty []byte
		pretty, _ = json.MarshalIndent(raw, "", "  ")
		fmt.Fprintf(&logs, "%s\n", string(pretty))
		fmt.Fprintf(&logs, "# download links:\n")
		for _, b := range books {
			fmt.Fprintf(&logs, "https://books.ms/main/%s\n", b.MD5)
		}
		fmt.Fprintf(&logs, "\n")
	}

	// Generate summary as text
	fmt.Fprintf(&logs, "Found %d books matching '%s':\n\n", len(books), query)
	for _, b := range books {
		fmt.Fprintf(&logs, "%-7s  %s — %s (%s) [%s]\n",
			b.ID, b.Title, b.Author, b.Year, b.MD5)
	}

	result.Logs = logs.String()
	return result, nil
}

// searchPapers searches for scientific papers using CrossRef API
func searchPapers(query string, limit int, logs *bytes.Buffer) (SearchResult, error) {
	result := SearchResult{}
	searchURL := fmt.Sprintf("https://api.crossref.org/works?query=%s&rows=%d",
		urlQueryEscape(query), limit)

	resp, err := http.Get(searchURL)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return result, fmt.Errorf("failed to search CrossRef: status %s, body: %s", resp.Status, string(bodyBytes))
	}

	var rawBytes []byte
	rawBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	if Debug {
		fmt.Fprintf(logs, "# crossref URL:\n%s\n", searchURL)
		fmt.Fprintf(logs, "# raw JSON:\n")
		var pretty []byte
		pretty, _ = json.MarshalIndent(json.RawMessage(rawBytes), "", "  ")
		fmt.Fprintf(logs, "%s\n", string(pretty))
	}

	var parsed struct {
		Message struct {
			Items []struct {
				DOI   string   `json:"DOI"`
				Title []string `json:"title"`
				Year  struct {
					DateParts [][]int `json:"date-parts"`
				} `json:"issued"`
				Author []struct {
					Family string `json:"family"`
					Given  string `json:"given"`
				} `json:"author"`
				Container struct {
					Title string `json:"title"`
				} `json:"container-title"`
			} `json:"items"`
		} `json:"message"`
	}

	if err := json.Unmarshal(rawBytes, &parsed); err != nil {
		return result, err
	}

	papers := make([]Paper, 0, len(parsed.Message.Items))
	for _, item := range parsed.Message.Items {
		if len(item.DOI) == 0 {
			continue
		}

		title := ""
		if len(item.Title) > 0 {
			title = item.Title[0]
		}

		author := ""
		if len(item.Author) > 0 {
			author = item.Author[0].Family
			if item.Author[0].Given != "" {
				author = item.Author[0].Given + " " + author
			}
		}

		year := ""
		if len(item.Year.DateParts) > 0 && len(item.Year.DateParts[0]) > 0 {
			year = fmt.Sprintf("%d", item.Year.DateParts[0][0])
		}

		journal := ""
		if item.Container.Title != "" {
			journal = item.Container.Title
		}

		// Fetch OA link via Unpaywall
		pdfLink := ""
		if item.DOI != "" {
			upURL := fmt.Sprintf("https://api.unpaywall.org/v2/%s?email=api@example.com", item.DOI)
			upResp, err := http.Get(upURL)

			if err == nil && upResp.StatusCode == http.StatusOK {
				defer upResp.Body.Close()
				var unpaywall struct {
					BestOALocation struct {
						URLForPDF string `json:"url_for_pdf"`
					} `json:"best_oa_location"`
				}

				upBytes, _ := io.ReadAll(upResp.Body)
				if Debug {
					fmt.Fprintf(logs, "# unpaywall URL for %s:\n%s\n", item.DOI, upURL)
				}

				if err := json.Unmarshal(upBytes, &unpaywall); err == nil && unpaywall.BestOALocation.URLForPDF != "" {
					pdfLink = unpaywall.BestOALocation.URLForPDF
					if Debug {
						fmt.Fprintf(logs, "# OA PDF link found: %s\n", pdfLink)
					}
				}
			}
		}

		papers = append(papers, Paper{
			DOI:     item.DOI,
			Title:   title,
			Author:  author,
			Year:    year,
			Journal: journal,
			PDFLink: pdfLink,
		})
	}

	result.Papers = papers

	// Generate summary as text
	fmt.Fprintf(logs, "Found %d papers matching '%s':\n\n", len(papers), query)
	for _, p := range papers {
		fmt.Fprintf(logs, "%s  %s — %s (%s)\n",
			p.DOI, p.Title, p.Author, p.Year)
		if p.PDFLink != "" {
			fmt.Fprintf(logs, "   ↳ %s\n", p.PDFLink)
		}
	}

	result.Logs = logs.String()
	return result, nil
}

/* ---------- download flow ---------- */

// DownloadBook downloads a book by its ID or a paper by its DOI
func DownloadBook(id, outPath, mirror string, scientific bool) (DownloadResult, error) {
	var logs bytes.Buffer
	result := DownloadResult{}

	if scientific {
		return downloadPaper(id, outPath, &logs)
	}

	if mirror == "" {
		mirror = DefaultMirror
	}

	_, _, books, err := fetchDetails([]string{id}, mirror)
	if err != nil {
		return result, err
	}
	if len(books) == 0 {
		return result, fmt.Errorf("no book found for id %s", id)
	}
	b := books[0]
	initialURL := fmt.Sprintf("https://books.ms/main/%s", b.MD5)

	if Debug {
		fmt.Fprintf(&logs, "# initial book page URL:\n%s\n", initialURL)
	}

	// Fetch the initial page to find the actual download link
	resp, err := http.Get(initialURL)
	if err != nil {
		return result, fmt.Errorf("failed to get initial book page %s: %w", initialURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body) // Read body for debugging if possible
		return result, fmt.Errorf("failed to get initial book page %s: status %s, body: %s", initialURL, resp.Status, string(bodyBytes))
	}

	pageDoc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return result, fmt.Errorf("failed to parse HTML from %s: %w", initialURL, err)
	}

	var foundHref string
	// Look for a link that points to download.books.ms and has a path structure like /main/PART1/PART2/FILENAME
	pageDoc.Find("a[href]").EachWithBreak(func(i int, s *goquery.Selection) bool {
		href, exists := s.Attr("href")
		if !exists {
			return true // continue
		}

		parsedHref, err_parse := url.Parse(href)
		if err_parse != nil {
			if Debug {
				fmt.Fprintf(&logs, "# skipping malformed href: %s, error: %v\n", href, err_parse)
			}
			return true // continue, malformed href
		}

		// Check if the host is download.books.ms or if it's a relative link that contains "download.books.ms/main/"
		hostIsDownloadBooksMs := parsedHref.Host == "download.books.ms"
		isLikelyRelativeDownloadLink := parsedHref.Host == "" && strings.Contains(href, "download.books.ms/main/")

		if hostIsDownloadBooksMs || isLikelyRelativeDownloadLink {
			pathTrimmed := strings.Trim(parsedHref.Path, "/")
			pathParts := strings.Split(pathTrimmed, "/")

			mainIndex := -1
			for idx, part := range pathParts {
				if part == "main" {
					mainIndex = idx
					break
				}
			}

			// We expect a structure like /main/PART1/PART2/FILENAME.EXT
			if mainIndex != -1 && len(pathParts) > mainIndex+3 {
				foundHref = href
				if Debug {
					fmt.Fprintf(&logs, "# matched download href: %s with host '%s' and path '%s'\n", href, parsedHref.Host, parsedHref.Path)
				}
				return false // stop searching
			} else if Debug && mainIndex != -1 {
				fmt.Fprintf(&logs, "# href '%s' matched host/path prefix but not depth: pathParts len %d, mainIndex %d\n", href, len(pathParts), mainIndex)
			}
		}
		return true // continue searching
	})

	if foundHref == "" {
		return result, fmt.Errorf("could not find a suitable download link on page %s. Looked for links to 'download.books.ms/main/' with at least 3 path segments after 'main'", initialURL)
	}

	// Resolve the found href against the page's URL (after redirects)
	pageFinalURL := resp.Request.URL
	actualDownloadURL, err := pageFinalURL.Parse(foundHref)
	if err != nil {
		return result, fmt.Errorf("failed to resolve download link '%s' against page URL '%s': %w", foundHref, pageFinalURL.String(), err)
	}

	actualDownloadURLString := actualDownloadURL.String()

	if Debug {
		fmt.Fprintf(&logs, "# found raw download href: %s\n", foundHref)
		fmt.Fprintf(&logs, "# page final URL for resolving relative links: %s\n", pageFinalURL.String())
		fmt.Fprintf(&logs, "# resolved actual download URL:\n%s\n", actualDownloadURLString)
	}

	out := outPath
	if out == "" {
		pathSegments := strings.Split(actualDownloadURL.Path, "/")
		if len(pathSegments) > 0 {
			filename := pathSegments[len(pathSegments)-1]
			decodedFilename, err_decode := url.PathUnescape(filename)
			if err_decode == nil && decodedFilename != "" {
				out = decodedFilename
			} else {
				out = b.MD5 + "." + b.Extension
				if Debug && err_decode != nil {
					fmt.Fprintf(&logs, "# failed to decode filename '%s' from URL path: %v. Falling back to MD5.extension\n", filename, err_decode)
				} else if Debug && decodedFilename == "" {
					fmt.Fprintf(&logs, "# filename '%s' from URL path is empty after decoding. Falling back to MD5.extension\n", filename)
				}
			}
		} else {
			out = b.MD5 + "." + b.Extension
			if Debug {
				fmt.Fprintf(&logs, "# URL path has no segments to extract filename. Falling back to MD5.extension\n")
			}
		}
		if Debug {
			fmt.Fprintf(&logs, "# determined output filename: %s\n", out)
		}
	}

	size, err := downloadFile(actualDownloadURLString, out, &logs)
	if err != nil {
		return result, err
	}

	result.FilePath = out
	result.Size = size
	result.Logs = logs.String()

	return result, nil
}

// downloadPaper downloads a scientific paper using its DOI
func downloadPaper(doi, outPath string, logs *bytes.Buffer) (DownloadResult, error) {
	result := DownloadResult{}

	// Use Unpaywall to find an open access PDF link
	upURL := fmt.Sprintf("https://api.unpaywall.org/v2/%s?email=api@example.com", doi)
	if Debug {
		fmt.Fprintf(logs, "# unpaywall URL:\n%s\n", upURL)
	}

	resp, err := http.Get(upURL)
	if err != nil {
		return result, fmt.Errorf("failed to query Unpaywall for DOI %s: %w", doi, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return result, fmt.Errorf("failed to query Unpaywall: status %s, body: %s", resp.Status, string(bodyBytes))
	}

	var unpaywall struct {
		BestOALocation struct {
			URLForPDF string `json:"url_for_pdf"`
		} `json:"best_oa_location"`
		Title string `json:"title"`
		DOI   string `json:"doi"`
	}

	upBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	if Debug {
		fmt.Fprintf(logs, "# raw JSON:\n")
		var pretty []byte
		pretty, _ = json.MarshalIndent(json.RawMessage(upBytes), "", "  ")
		fmt.Fprintf(logs, "%s\n", string(pretty))
	}

	if err := json.Unmarshal(upBytes, &unpaywall); err != nil {
		return result, fmt.Errorf("failed to parse Unpaywall response: %w", err)
	}

	if unpaywall.BestOALocation.URLForPDF == "" {
		return result, fmt.Errorf("no open-access PDF found for DOI: %s", doi)
	}

	pdfURL := unpaywall.BestOALocation.URLForPDF
	if Debug {
		fmt.Fprintf(logs, "# PDF URL:\n%s\n", pdfURL)
	}

	out := outPath
	if out == "" {
		// Create a filename from the DOI if outPath is not provided
		safeDOI := strings.ReplaceAll(doi, "/", "_")
		out = safeDOI + ".pdf"

		// If we have a title, use it to create a better filename
		if unpaywall.Title != "" {
			safeTitle := sanitizeFilename(unpaywall.Title)
			if safeTitle != "" {
				out = safeTitle + ".pdf"
			}
		}

		if Debug {
			fmt.Fprintf(logs, "# determined output filename: %s\n", out)
		}
	}

	size, err := downloadFile(pdfURL, out, logs)
	if err != nil {
		return result, err
	}

	result.FilePath = out
	result.Size = size
	result.Logs = logs.String()

	return result, nil
}

/* ---------- libgen helpers ---------- */

// crawlIDs fetches book IDs matching the search query
func crawlIDs(query string, limit int, mirror string) (string, []string, error) {
	v := url.Values{
		"req":    {query},
		"res":    {fmt.Sprint(limit)},
		"view":   {"simple"},
		"phrase": {"1"},
		"column": {"def"},
	}
	searchURL := fmt.Sprintf("%s/search.php?%s", mirror, v.Encode())

	doc, err := goquery.NewDocument(searchURL)
	if err != nil {
		return "", nil, err
	}
	var ids []string
	doc.Find("table.c tr").Each(func(i int, s *goquery.Selection) {
		if i == 0 || len(ids) >= limit {
			return
		}
		if id := strings.TrimSpace(s.Find("td").First().Text()); id != "" {
			ids = append(ids, id)
		}
	})
	return searchURL, ids, nil
}

// fetchDetails fetches detailed book information by IDs
func fetchDetails(ids []string, mirror string) (string, []map[string]string, []Book, error) {
	fields := "id,title,author,year,extension,filesize,md5,pages"
	jsonURL := fmt.Sprintf("%s/json.php?object=libgen&ids=%s&fields=%s",
		mirror, strings.Join(ids, ","), fields)

	resp, err := http.Get(jsonURL)
	if err != nil {
		return "", nil, nil, err
	}
	defer resp.Body.Close()

	var raw []map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return "", nil, nil, err
	}

	books := make([]Book, 0, len(raw))
	for _, m := range raw {
		books = append(books, Book{
			ID:        m["id"],
			Title:     m["title"],
			Author:    m["author"],
			Year:      m["year"],
			Pages:     m["pages"],
			Extension: m["extension"],
			MD5:       m["md5"],
			FileSize:  m["filesize"],
		})
	}
	return jsonURL, raw, books, nil
}

// downloadFile downloads a file from the URL to the specified path
func downloadFile(url, out string, logs *bytes.Buffer) (int64, error) {
	log.Info().Str("url", url).Str("out", out).Msg("Downloading file")
	fmt.Fprintf(logs, "Downloading file from: %s to: %s\n", url, out)

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("failed to download file: status %s, body: %s", resp.Status, string(bodyBytes))
	}

	f, err := os.Create(out)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	size, err := io.Copy(f, resp.Body)
	if err != nil {
		return 0, err
	}

	log.Info().
		Str("path", out).
		Int64("size", size).
		Msg("File downloaded successfully")

	fmt.Fprintf(logs, "Downloaded %s (%.2f MB) to %s\n",
		out, float64(size)/(1024*1024), out)

	return size, nil
}

// PrettyPrintBooks prints a summary of books to the provided writer
func PrettyPrintBooks(books []Book, w io.Writer) {
	for _, b := range books {
		fmt.Fprintf(w, "%-7s  %s — %s (%s) [%s]\n",
			b.ID, b.Title, b.Author, b.Year, b.MD5)
	}
}

// PrettyPrintPapers prints a summary of papers to the provided writer
func PrettyPrintPapers(papers []Paper, w io.Writer) {
	for _, p := range papers {
		fmt.Fprintf(w, "%s  %s — %s (%s)\n",
			p.DOI, p.Title, p.Author, p.Year)
		if p.PDFLink != "" {
			fmt.Fprintf(w, "   ↳ %s\n", p.PDFLink)
		}
	}
}

// Helper functions for scientific paper handling
func urlQueryEscape(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}

// sanitizeFilename creates a safe filename from a title
func sanitizeFilename(title string) string {
	// Remove characters that are problematic in filenames
	forbidden := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := title
	for _, char := range forbidden {
		result = strings.ReplaceAll(result, char, "_")
	}

	// Limit length to avoid excessively long filenames
	if len(result) > 100 {
		result = result[:100]
	}

	return strings.TrimSpace(result)
}
