package main

import (
	"fmt"
	"image/jpeg"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jdeng/goheif"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

// Skip Writer for exif writing
type writerSkipper struct {
	w           io.Writer
	bytesToSkip int
}

func (w *writerSkipper) Write(data []byte) (int, error) {
	if w.bytesToSkip <= 0 {
		return w.w.Write(data)
	}

	if dataLen := len(data); dataLen < w.bytesToSkip {
		w.bytesToSkip -= dataLen
		return dataLen, nil
	}

	if n, err := w.w.Write(data[w.bytesToSkip:]); err == nil {
		n += w.bytesToSkip
		w.bytesToSkip = 0
		return n, nil
	} else {
		return n, err
	}
}

func newWriterExif(w io.Writer, exif []byte) (io.Writer, error) {
	writer := &writerSkipper{w, 2}
	soi := []byte{0xff, 0xd8}
	if _, err := w.Write(soi); err != nil {
		return nil, err
	}

	if exif != nil {
		app1Marker := 0xe1
		markerlen := 2 + len(exif)
		marker := []byte{0xff, uint8(app1Marker), uint8(markerlen >> 8), uint8(markerlen & 0xff)}
		if _, err := w.Write(marker); err != nil {
			return nil, err
		}

		if _, err := w.Write(exif); err != nil {
			return nil, err
		}
	}

	return writer, nil
}

func convert(heicFilePath, outJpgPath string) error {

	fin, fout := heicFilePath, outJpgPath
	fi, err := os.Open(fin)
	if err != nil {
		return err
	}
	defer fi.Close()

	exif, err := goheif.ExtractExif(fi)
	if err != nil {
		return err
	}

	img, err := goheif.Decode(fi)
	if err != nil {
		return err
	}

	fo, err := os.OpenFile(fout, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fo.Close()

	w, _ := newWriterExif(fo, exif)
	err = jpeg.Encode(w, img, nil)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	file, _ := exec.LookPath(os.Args[0])
	dir, _ := path.Split(file)
	os.Chdir(dir)
	wd, _ := os.Getwd()
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	p2 := widgets.NewParagraph()
	p2.Text = "工作路径: " + wd
	p2.Border = false
	p2.SetRect(0, 0, 75, 6)
	p2.TextStyle.Fg = ui.ColorGreen

	p := widgets.NewParagraph()
	p.Title = "消息"
	p.Text = ""
	p.SetRect(0, 3, 60, 8)
	p.TextStyle.Fg = ui.ColorWhite
	p.BorderStyle.Fg = ui.ColorCyan

	g4 := widgets.NewGauge()
	g4.Title = "进度"
	g4.SetRect(0, 8, 60, 13)
	g4.Percent = 0
	g4.Label = ""
	g4.BarColor = ui.ColorGreen
	g4.LabelStyle = ui.NewStyle(ui.ColorYellow)

	listData := []string{}
	l := widgets.NewList()
	l.Title = "转换失败"
	l.Rows = listData
	l.SetRect(0, 14, 60, 25)
	l.TextStyle.Fg = ui.ColorYellow

	ui.Render(p2, g4, p, l)

	heicFilePath := []string{}
	filepath.Walk(wd, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if strings.HasSuffix(info.Name(), ".heic") {
				heicFilePath = append(heicFilePath, path)
				p.Text = "找到文件: " + path[len(wd)+1:]
				g4.Label = "查找文件，已找到" + strconv.Itoa(len(heicFilePath)) + "个文件"
				if len(heicFilePath)%10 == 0 {
					g4.Percent += 1
				}
				ui.Render(p, g4)
			}
		}
		return nil
	})
	p.Text = "查找完成，共找到 " + strconv.Itoa(len(heicFilePath)) + "个文件"
	g4.Label = "查找完成"
	g4.Percent = 100
	ui.Render(p, g4)
	time.Sleep(1 * time.Second)

	p.Text = "开始将heic文件转jpg"
	g4.Label = "开始转换"
	g4.Percent = 0
	ui.Render(p, g4)
	time.Sleep(1 * time.Second)

	convCount := 0
	for _, r := range heicFilePath {
		convCount++
		p.Text = "转换文件: " + r[len(wd)+1:]
		g4.Label = fmt.Sprintf("转换进度，(%d/%d)", convCount, len(heicFilePath))
		g4.Percent = int((float32(convCount) / float32(len(heicFilePath))) * 100)
		ui.Render(p, g4)
		output := r[:len(r)-5] + ".jpg"
		err := convert(r, output)
		if err != nil {
			listData = append(listData, r[len(wd)+1:])
			l.Title = fmt.Sprintf("转换失败 (%d)", len(listData))
			l.Rows = listData
			ui.Render(l)
		}
	}

	p.Text = fmt.Sprintf("转换完成，共转换文件 %d 个，失败 %d。\n按q键退出。", len(heicFilePath), len(listData))
	ui.Render(p)

	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		}
	}

}
