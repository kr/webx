// Usage: webxd start addr path
// Params:
//   addr - inner addr to send requests to
//   path - path to execute child
// Environment:
//   PORT     - outer port to listen
//   WEBX_URL - location and credentials for RSPDY connection
//              https://foo@route.webx.io/
package main

var tlsConfig = &tls.Config{
	InsecureSkipVerify: true,
	NextProtos:         []string{"rspdy/3"},
}

// Params for "child" subcommand:
//  addr - same as for start
//  wfd  - fd to signal readiness to parent
func main() {
	log.SetFlags(0)
	log.SetPrefix("webxd: ")
	switch os.Args[1] {
	case "start":
		start(os.Args[2], os.Args[3])
	case "child":
		child(os.Args[2], os.Args[3])
	}
}

func start(innerAddr, path string) {
	r, w, err := os.Pipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd := exec.Command("webxd", "child", innerAddr, strconv.Itoa(w.Fd()))
	cmd.Path = path
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = []*os.File{w}
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	w.Close()
	n, _ := r.Read(make([]byte, 1))
	if n < 1 {
		os.Exit(1) // Child failed; child logs the error message directly.
	}
}

func child(innerAddr, wfd string) {
	wfdn, err := strconv.Atoi(wfd)
	if err != nil {
		log.Fatal(err)
	}
	parentPipe := os.NewFile(wfdn, "|parent")

	routerURL, err := url.Parse(os.Getenv("WEBX_URL"))
	if err != nil {
		log.Fatal("parse url:", err)
	}
	mustSanityCheckURL(routerURL)

	handshake := func(w http.ResponseWriter, r *http.Request) {
		n, err := parentPipe.Write([]byte{0})
		if n < 1 {
			log.Fatal("signal parent:", err)
		}
		parentPipe.Close()

		webxName := routerURL.User.Username()
		cmd := BackendCommand{"add", webxName}
		err = json.NewEncoder(w).Encode(cmd)
		if err != nil {
			log.Fatal("handshake:", err)
		}
		select {}
	}

	innerURL := &url.URL{Scheme: "http", Host: innerAddr}
	http.Handle("/", httputil.NewSingleHostReverseProxy(innerURL))
	http.HandleFunc("backend.webx.io/names", handshake)
	s, err := rspdy.DialAndServe(routerURL.Host, tlsConfig, nil)
	if err != nil {
		log.Fatal("DialAndServe:", err)
	}
}

func mustSanityCheckURL(u *url.URL) {
	if u.User == nil {
		log.Fatal("url has no userinfo")
	}
	if u.Scheme != "https" {
		log.Fatal("scheme must be https")
	}
	if u.Path != "/" {
		log.Fatal("path must be /")
	}
	if u.RawQuery != "" {
		log.Fatal("query must be empty")
	}
	if u.Fragment != "" {
		log.Fatal("fragment must be empty")
	}
}

type BackendCommand struct {
	Op   string // "add" or "remove"
	Name string // e.g. "foo" for foo.webxapp.io
}
