package efsu

import (
	"io/fs"
	"time"
)

const (
	CommandList     = "List"
	CommandDownload = "Download"
)

type Input struct {
	Command  string
	Version  string
	YOLO     bool
	List     *ListInput     `json:",omitempty"`
	Download *DownloadInput `json:",omitempty"`
}

type Output struct {
	Version  string
	List     *ListOutput     `json:",omitempty"`
	Download *DownloadOutput `json:",omitempty"`
}

type ListInput struct {
	Path      string
	Recursive bool
	Next      string `json:",omitempty"`
}

type ListOutput struct {
	Items []ListItem
	Next  string `json:",omitempty"`
}

type ListItem struct {
	Path    string
	Mode    fs.FileMode
	Size    int64
	ModTime time.Time
}

type DownloadInput struct {
	Path   string
	Offset int64
}

type DownloadOutput struct {
	FileSize   int64
	Mode       fs.FileMode
	Content    []byte
	NextOffset int64
}
