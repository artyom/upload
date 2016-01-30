// Command upload uploads files to new directory on a remote server via ssh. For
// each command call new randomly-named remote subdirectory is created.
package main

import (
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/artyom/autoflags"
	"github.com/pkg/browser"
	"github.com/pkg/sftp"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func main() {
	params := struct {
		User string `flag:"user,ssh connection username"`
		Addr string `flag:"addr,ssh host:port"`
		Dir  string `flag:"dir,remote directory to upload files to"`
		Url  string `flag:"url,remote url base to open after upload"`
		Long bool   `flag:"long,generate long subdirectory name"`
	}{
		User: os.Getenv("USER"),
		Addr: "localhost:22",
		Dir:  "/tmp",
	}
	autoflags.Define(&params)
	flag.Parse()

	cfg, err := config(params.User)
	if err != nil {
		log.Fatal(err)
	}
	res, err := upload(params.Addr, params.Dir, params.Long, cfg, flag.Args())
	if res != "" {
		fmt.Println(res)
		if params.Url != "" {
			browser.OpenURL(path.Join(params.Url, path.Base(res)))
		}
	}
	if err != nil {
		log.Fatal(err)
	}
}

func upload(addr, dir string, long bool, config *ssh.ClientConfig, files []string) (string, error) {
	if len(files) == 0 {
		return "", errors.New("nothing to upload")
	}
	for _, name := range files {
		if fi, err := os.Stat(name); err != nil {
			return "", err
		} else if fi.IsDir() {
			return "", errors.New("upload of directories not yet implemented")
		}
	}
	conn, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	sc, err := sftp.NewClient(conn)
	if err != nil {
		return "", err
	}
	defer sc.Close()
	if di, err := sc.Stat(dir); err != nil {
		return "", err
	} else if !di.IsDir() {
		return "", errors.New("destination path is not a directory")
	}
	var dst string
	if long {
		for i := 0; i < 3; i++ {
			dst = path.Join(dir, fmt.Sprintf("%x", randBytes()))
			if err = sc.Mkdir(dst); err == nil {
				break
			}
		}
	} else {
		for b, i := randBytes(), 1; i < len(b); i++ {
			dst = path.Join(dir, fmt.Sprintf("%x", b[:i]))
			if err = sc.Mkdir(dst); err == nil {
				break
			}
		}
	}
	if err != nil {
		return "", errors.New("failed to create new random-named directory, probably too many already exist")
	}

	for _, name := range files {
		if err := uploadFile(name, dst, sc); err != nil {
			return dst, err
		}
	}
	return dst, nil
}

func uploadFile(name, dir string, sc *sftp.Client) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	of, err := sc.Create(path.Join(dir, filepath.Base(name)))
	if err != nil {
		return err
	}
	defer of.Close()
	bufp := copyBufPool.Get().(*[]byte)
	defer copyBufPool.Put(bufp)
	if _, err := io.CopyBuffer(of, f, *bufp); err != nil {
		return err
	}
	return of.Close()
}

func randBytes() []byte {
	b := make([]byte, 10)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return b
}

func config(user string) (*ssh.ClientConfig, error) {
	agentConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, err
	}
	// not closing agentConn here because it should be left open for use in
	// ssh.ClientConfig
	sshAgent := agent.NewClient(agentConn)
	signers, err := sshAgent.Signers()
	if err != nil {
		agentConn.Close()
		return nil, err
	}
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signers...)},
	}
	return config, nil
}

func init() {
	log.SetFlags(0)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] file...\n", os.Args[0])
		flag.PrintDefaults()
	}
}

var copyBufPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 256*1024)
		return &b
	},
}
