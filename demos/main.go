package main

import (
	//"archive/tar"
	"archive/zip"
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"html/template"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

import "github.com/JamesDunne/go-util/base"
import "github.com/JamesDunne/go-util/web"

var proxyRoot, jailRoot, accelRedirect string
var jplayerUrl, jplayerPath string
var useJPlayer bool

var uiTmpl *template.Template

func removeIfStartsWith(s, start string) string {
	if !strings.HasPrefix(s, start) {
		return s
	}
	return s[len(start):]
}

func translateForProxy(s string) string {
	return path.Join(proxyRoot, removeIfStartsWith(s, jailRoot))
}

// For directory entry sorting:

type Entries []os.FileInfo

func (s Entries) Len() int      { return len(s) }
func (s Entries) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type sortBy int

const (
	sortByName sortBy = iota
	sortByDate
	sortBySize
)

type sortDirection int

const (
	sortAscending sortDirection = iota
	sortDescending
)

// Sort by name:
type ByName struct {
	Entries
	dir sortDirection
}

func (s ByName) Less(i, j int) bool {
	if s.Entries[i].IsDir() && !s.Entries[j].IsDir() {
		return true
	}
	if !s.Entries[i].IsDir() && s.Entries[j].IsDir() {
		return false
	}

	if s.dir == sortAscending {
		return s.Entries[i].Name() < s.Entries[j].Name()
	} else {
		return s.Entries[i].Name() > s.Entries[j].Name()
	}
}

// Sort by last modified time:
type ByDate struct {
	Entries
	dir sortDirection
}

func (s ByDate) Less(i, j int) bool {
	if s.Entries[i].IsDir() && !s.Entries[j].IsDir() {
		return true
	}
	if !s.Entries[i].IsDir() && s.Entries[j].IsDir() {
		return false
	}

	if s.dir == sortAscending {
		return s.Entries[i].ModTime().Before(s.Entries[j].ModTime())
	} else {
		return s.Entries[i].ModTime().After(s.Entries[j].ModTime())
	}
}

// Sort by size:
type BySize struct {
	Entries
	dir sortDirection
}

func (s BySize) Less(i, j int) bool {
	if s.Entries[i].IsDir() && !s.Entries[j].IsDir() {
		return true
	}
	if !s.Entries[i].IsDir() && s.Entries[j].IsDir() {
		return false
	}

	if s.dir == sortAscending {
		return s.Entries[i].Size() < s.Entries[j].Size()
	} else {
		return s.Entries[i].Size() > s.Entries[j].Size()
	}
}

func followSymlink(localPath string, dfi os.FileInfo) os.FileInfo {
	// Check symlink:
	if (dfi.Mode() & os.ModeSymlink) != 0 {

		dfiPath := path.Join(localPath, dfi.Name())
		if targetPath, err := os.Readlink(dfiPath); err == nil {
			// Find the absolute path of the symlink's target:
			if !path.IsAbs(targetPath) {
				targetPath = path.Join(localPath, targetPath)
			}
			if tdfi, err := os.Stat(targetPath); err == nil {
				// Change to the target so we get its properties instead of the symlink's:
				return tdfi
			}
		}
	}

	return dfi
}

// Logging+action functions
func doError(req *http.Request, rsp http.ResponseWriter, msg string, code int) {
	http.Error(rsp, msg, code)
}

func doRedirect(req *http.Request, rsp http.ResponseWriter, url string, code int) {
	http.Redirect(rsp, req, url, code)
}

func doOK(req *http.Request, msg string, code int) {
}

// Marshal an object to JSON or panic.
func marshal(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func isMP3(filename string) bool {
	ext := path.Ext(filename)
	mt := mime.TypeByExtension(ext)
	if mt != "audio/mpeg" {
		return false
	}
	if ext != ".mp3" {
		return false
	}
	return true
}

type jplayerMP3Entry struct {
	Title string `json:"title"`
	MP3   string `json:"mp3"`
}

type indexEntry struct {
	HREF     string
	Name     string
	ModTime  time.Time // .Format("2006-01-02 15:04:05 -0700 MST")
	Size     template.HTML
	MimeType string
}

func generateIndexHtml(rsp http.ResponseWriter, req *http.Request, u *url.URL) {
	// Build index.html
	relPath := removeIfStartsWith(u.Path, proxyRoot)

	localPath := path.Join(jailRoot, relPath)
	pathLink := path.Join(proxyRoot, relPath)

	baseDir := path.Dir(localPath)
	if localPath[len(localPath)-1] == '/' {
		baseDir = path.Dir(localPath[0 : len(localPath)-1])
	}
	if baseDir == "" {
		baseDir = "/"
	}

	// Determine what mode to sort by...
	sortString := ""

	// Check the .index-sort file:
	if sf, err := os.Open(path.Join(localPath, ".index-sort")); err == nil {
		defer sf.Close()
		scanner := bufio.NewScanner(sf)
		if scanner.Scan() {
			sortString = scanner.Text()
		}
	}

	// Use query-string 'sort' to override sorting:
	sortStringQuery := u.Query().Get("sort")
	if sortStringQuery != "" {
		sortString = sortStringQuery
	}

	// default Sort mode for headers
	nameSort := "name-asc"
	dateSort := "date-asc"
	sizeSort := "size-asc"

	// Determine the sorting mode:
	sortBy, sortDir := sortByName, sortAscending
	switch sortString {
	case "size-desc":
		sortBy, sortDir = sortBySize, sortDescending
	case "size-asc":
		sortBy, sortDir = sortBySize, sortAscending
		sizeSort = "size-desc"
	case "date-desc":
		sortBy, sortDir = sortByDate, sortDescending
	case "date-asc":
		sortBy, sortDir = sortByDate, sortAscending
		dateSort = "date-desc"
	case "name-desc":
		sortBy, sortDir = sortByName, sortDescending
	case "name-asc":
		sortBy, sortDir = sortByName, sortAscending
		nameSort = "name-desc"
	default:
	}

	// Open the directory to read its contents:
	f, err := os.Open(localPath)
	if err != nil {
		doError(req, rsp, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	// Read the directory entries:
	fis, err := f.Readdir(0)
	if err != nil {
		doError(req, rsp, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if there are MP3s in this directory:
	hasMP3s := false
	if useJPlayer {
		for _, dfi := range fis {
			dfi = followSymlink(localPath, dfi)
			if !isMP3(dfi.Name()) {
				continue
			}
			hasMP3s = true
			break
		}
	}

	// Sort the entries by the desired mode:
	switch sortBy {
	default:
		sort.Sort(ByName{fis, sortDir})
	case sortByName:
		sort.Sort(ByName{fis, sortDir})
	case sortByDate:
		sort.Sort(ByDate{fis, sortDir})
	case sortBySize:
		sort.Sort(BySize{fis, sortDir})
	}

	// TODO: check Accepts header to reply accordingly (i.e. add JSON support)

	model := &struct {
		Title      string
		PathHREF   string
		ParentHREF string
		SortName   string
		SortDate   string
		SortSize   string
		Entries    []indexEntry

		HasMP3s     bool
		JPlayerURL  string
		JPlayerMP3s []jplayerMP3Entry
		//JPlayerJSON template.JS
	}{
		Title:      html.EscapeString(pathLink),
		PathHREF:   html.EscapeString(pathLink), // TODO: URL path escape here!
		ParentHREF: "",
		SortName:   nameSort,
		SortDate:   dateSort,
		SortSize:   sizeSort,

		HasMP3s:    hasMP3s,
		JPlayerURL: jplayerUrl,
		//JPlayerJSON: template.JS(""),
	}

	// Add the Parent Directory link if we're above the jail root:
	if strings.HasPrefix(baseDir, jailRoot) {
		model.ParentHREF = translateForProxy(baseDir) + "/"
	}

	if hasMP3s {
		// Generate jPlayer playlist:
		model.JPlayerMP3s = make([]jplayerMP3Entry, 0, len(fis))

		for _, dfi := range fis {
			name := dfi.Name()
			if name[0] == '.' {
				continue
			}

			dfi = followSymlink(localPath, dfi)

			dfiPath := path.Join(localPath, name)
			href := translateForProxy(dfiPath)

			if dfi.IsDir() {
				continue
			}

			if !isMP3(name) {
				continue
			}

			ext := path.Ext(name)
			onlyname := name
			if ext != "" {
				onlyname = name[0 : len(name)-len(ext)]
			}

			model.JPlayerMP3s = append(model.JPlayerMP3s, jplayerMP3Entry{
				Title: onlyname,
				MP3:   href,
			})
		}

		//model.JPlayerJSON = template.JS(marshal(mp3List))
	}

	model.Entries = make([]indexEntry, 0, len(fis))
	for _, dfi := range fis {
		name := dfi.Name()
		if name[0] == '.' {
			continue
		}

		dfiPath := path.Join(localPath, name)
		dfi = followSymlink(localPath, dfi)

		href := translateForProxy(dfiPath)
		mt := mime.TypeByExtension(path.Ext(dfi.Name()))

		sizeText := ""
		if dfi.IsDir() {
			sizeText = "-"
			name += "/"
			href += "/"
		} else {
			size := dfi.Size()
			if size < 1024 {
				sizeText = fmt.Sprintf("%d  B", size)
			} else if size < 1024*1024 {
				sizeText = fmt.Sprintf("%.02f KB", float64(size)/1024.0)
			} else if size < 1024*1024*1024 {
				sizeText = fmt.Sprintf("%.02f MB", float64(size)/(1024.0*1024.0))
			} else {
				sizeText = fmt.Sprintf("%.02f GB", float64(size)/(1024.0*1024.0*1024.0))
			}
		}

		model.Entries = append(model.Entries, indexEntry{
			HREF:     href,
			Name:     name,
			ModTime:  dfi.ModTime(),
			Size:     template.HTML(strings.Replace(html.EscapeString(sizeText), " ", "&nbsp;", -1)),
			MimeType: mt,
		})
	}

	rsp.Header().Add("Content-Type", "text/html; charset=utf-8")
	err = uiTmpl.ExecuteTemplate(rsp, "index", model)
	if err != nil {
		log.Printf("%s\n", err.Error())
		return
	}
	doOK(req, localPath, http.StatusOK)
	return
}

// Downloads contents of the current directory as a ZIP file, streamed to the client's browser:
func downloadZip(rsp http.ResponseWriter, req *http.Request, u *url.URL, dir *os.FileInfo, localPath string) {
	// Generate a decent filename based on the folder URL:
	fullName := removeIfStartsWith(localPath, jailRoot)
	fullName = removeIfStartsWith(fullName, "/")
	// Translate '/' separators into '-':
	fullName = strings.Map(func(i rune) rune {
		if i == '/' {
			return '-'
		} else {
			return i
		}
	}, fullName)

	var fis []os.FileInfo
	{
		// Open the directory to read its contents:
		df, err := os.Open(localPath)
		if err != nil {
			doError(req, rsp, err.Error(), http.StatusInternalServerError)
			return
		}
		defer df.Close()

		// Read the directory entries:
		fis, err = df.Readdir(0)
		if err != nil {
			doError(req, rsp, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Make sure filenames are in ascending order:
	sort.Sort(ByName{fis, sortAscending})

	// Start with a 200 status and set up the download:
	h := rsp.Header()
	h.Set("Content-Type", "application/zip")
	h.Set("Content-Description", "File Transfer")
	// NOTE(jsd): Need proper HTTP value encoding here!
	h.Set("Content-Disposition", "attachment; filename=\""+fullName+".zip\"")
	h.Set("Content-Transfer-Encoding", "binary")

	// Here we estimate the final length of the ZIP file being streamed:
	const (
		fileHeaderLen      = 30 // + filename + extra
		dataDescriptorLen  = 16 // four uint32: descriptor signature, crc32, compressed size, size
		directoryHeaderLen = 46 // + filename + extra + comment
		directoryEndLen    = 22 // + comment
	)

	zipLength := 0
	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}
		zipLength += fileHeaderLen
		zipLength += len(fi.Name())
		// + extra

		// TODO(jsd): ZIP64 support
		size := fi.Size()
		zipLength += int(size)
		zipLength += dataDescriptorLen

		// Directory entries:
		zipLength += directoryHeaderLen
		zipLength += len(fi.Name())
		// + extra
		// + comment
	}
	zipLength += directoryEndLen

	h.Set("Content-Length", fmt.Sprintf("%d", zipLength))

	rsp.WriteHeader(http.StatusOK)

	// Create a zip stream writing to the HTTP response:
	zw := zip.NewWriter(rsp)
	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}
		name := fi.Name()
		if name[0] == '.' {
			continue
		}

		// Dereference symlinks, if applicable:
		fi = followSymlink(localPath, fi)
		fiPath := path.Join(localPath, name)

		// Open the source file for reading:
		lf, err := os.Open(fiPath)
		if err != nil {
			panic(err)
		}
		defer lf.Close()

		// Create the ZIP entry to write to:
		zfh, err := zip.FileInfoHeader(fi)
		if err != nil {
			panic(err)
		}
		// Don't bother compressing the file:
		zfh.Method = zip.Store

		zf, err := zw.CreateHeader(zfh)
		if err != nil {
			panic(err)
		}

		// Copy the file contents into the ZIP:
		_, err = io.Copy(zf, lf)
		if err != nil {
			panic(err)
		}
	}

	// Mark the end of the ZIP stream:
	zw.Close()
	return
}

func processProxiedRequest(rsp http.ResponseWriter, req *http.Request, u *url.URL) {
	relPath := removeIfStartsWith(u.Path, proxyRoot)
	localPath := path.Join(jailRoot, relPath)

	// Check if the requested path is a symlink:
	fi, err := os.Lstat(localPath)
	if fi != nil && (fi.Mode()&os.ModeSymlink) != 0 {
		localDir := path.Dir(localPath)

		// Check if file is a symlink and do 302 redirect:
		linkDest, err := os.Readlink(localPath)
		if err != nil {
			doError(req, rsp, err.Error(), http.StatusBadRequest)
			return
		}

		// NOTE(jsd): Problem here for links outside the jail folder.
		if path.IsAbs(linkDest) && !strings.HasPrefix(linkDest, jailRoot) {
			doError(req, rsp, "Symlink points outside of jail", http.StatusBadRequest)
			return
		}

		linkDest = path.Join(localDir, linkDest)
		tp := translateForProxy(linkDest)

		doRedirect(req, rsp, tp, http.StatusFound)
		return
	}

	// Regular stat
	fi, err = os.Stat(localPath)
	if err != nil {
		doError(req, rsp, err.Error(), http.StatusNotFound)
		return
	}

	// Serve the file if it is regular:
	if fi.Mode().IsRegular() {
		// Send file:

		// NOTE(jsd): using `http.ServeFile` does not appear to handle range requests well. Lots of broken pipe errors
		// that lead to a poor client experience. X-Accel-Redirect back to nginx is much better.

		if accelRedirect != "" {
			// Use X-Accel-Redirect if the cmdline option was given:
			redirPath := path.Join(accelRedirect, relPath)
			rsp.Header().Add("X-Accel-Redirect", redirPath)
			rsp.Header().Add("Content-Type", mime.TypeByExtension(path.Ext(localPath)))
			rsp.WriteHeader(200)
		} else {
			// Just serve the file directly from the filesystem:
			http.ServeFile(rsp, req, localPath)
		}

		return
	}

	// Generate an index.html for directories:
	if fi.Mode().IsDir() {
		dl := u.Query().Get("dl")
		if dl != "" {
			switch dl {
			default:
				fallthrough
			case "zip":
				downloadZip(rsp, req, u, &fi, localPath)
				//case "tar":
				//	downloadTar(rsp, req, u, &fi, localPath)
			}
			return
		}

		generateIndexHtml(rsp, req, u)
		return
	}
}

// Serves an index.html file for a directory or sends the requested file.
func processRequest(rsp http.ResponseWriter, req *http.Request) {
	// proxy sends us absolute path URLs
	u, err := url.Parse(req.RequestURI)
	if err != nil {
		log.Fatal(err)
	}

	if (jplayerPath != "") && strings.HasPrefix(u.Path, jplayerUrl) {
		// URL is under the jPlayer path:
		localPath := path.Join(jplayerPath, removeIfStartsWith(u.Path, jplayerUrl))
		http.ServeFile(rsp, req, localPath)
		return
	} else if strings.HasPrefix(u.Path, proxyRoot) {
		// URL is under the proxy path:
		processProxiedRequest(rsp, req, u)
		return
	}
}

func main() {
	fl_listen_uri := flag.String("l", "tcp://0.0.0.0:8080", "listen URI (schemes available: tcp, unix)")
	flag.StringVar(&proxyRoot, "p", "/", "root of web requests to process")
	flag.StringVar(&jailRoot, "r", ".", "local filesystem path to bind to web request root path")
	htmlPath := flag.String("html", ".", "local filesystem path to HTML templates")
	flag.StringVar(&accelRedirect, "xa", "", "Root of X-Accel-Redirect paths to use)")
	flag.StringVar(&jplayerUrl, "jp-url", "", `Web path to jPlayer files (e.g. "/js" or "//static.somesite.com/jp")`)
	flag.StringVar(&jplayerPath, "jp-path", "", `Local filesystem path to serve jPlayer files from (optional)`)
	flag.Parse()

	// Parse all the URIs:
	listen_addr, err := base.ParseListenable(*fl_listen_uri)
	base.PanicIf(err)

	jailRoot = base.CanonicalPath(jailRoot)

	if jplayerUrl != "" {
		useJPlayer = true
	}

	// Watch the html templates for changes and reload them:
	_, cleanup, err := web.WatchTemplates("ui", *htmlPath, "*.html", func(t *template.Template) *template.Template {
		return t.Funcs(map[string]interface{}{
			"isLast": func(i, count int) bool { return i == count-1 },
		})
	}, &uiTmpl)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer cleanup()

	// Start the server:
	base.ServeMain(listen_addr, func(l net.Listener) error {
		return http.Serve(l, http.HandlerFunc(processRequest))
	})
}
