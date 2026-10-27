package app

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/getlantern/systray"
)

// SetupTray starts the system tray icon and event loop in a goroutine.
func SetupTray(a *App) {
	go systray.Run(func() {
		iconBytes := GenerateDefaultIcon()
		if len(iconBytes) > 0 {
			systray.SetIcon(iconBytes)
		}
		systray.SetTitle("FILE SYNC")
		systray.SetTooltip("FILE SYNC - 초경량 파일 동기화")

		mShow := systray.AddMenuItem("창 열기", "앱 창을 화면에 표시합니다")
		mHide := systray.AddMenuItem("백그라운드 숨기기", "창을 트레이로 숨깁니다")
		systray.AddSeparator()
		mToggleWatch := systray.AddMenuItem("감시 토글", "실시간 감시를 켜거나 끕니다")
		systray.AddSeparator()
		mQuit := systray.AddMenuItem("완전 종료", "프로그램을 종료합니다")

		for {
			select {
			case <-mShow.ClickedCh:
				a.ShowWindow()
			case <-mHide.ClickedCh:
				a.HideWindow()
			case <-mToggleWatch.ClickedCh:
				a.ToggleWatchFromTray()
			case <-mQuit.ClickedCh:
				systray.Quit()
				a.Quit()
				return
			}
		}
	}, func() {
	})
}

// GenerateDefaultIcon creates a standalone 16x16 ICO byte slice with PNG payload.
func GenerateDefaultIcon() []byte {
	const size = 16
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	cyan := color.RGBA{R: 6, G: 182, B: 212, A: 255}
	emerald := color.RGBA{R: 16, G: 185, B: 129, A: 255}

	pixelMap := []string{
		"................",
		".....CCCC.......",
		"...CC....CC.E...",
		"..C........EE...",
		".C.......EEEE...",
		".C.........EE...",
		"C...........E...",
		"C...............",
		"...............C",
		"...E...........C",
		"...EE.........C.",
		"...EEEE.......C.",
		"...EE........C..",
		"...E..CC....CC..",
		".......CCCC.....",
		"................",
	}

	for y, row := range pixelMap {
		for x := 0; x < len(row) && x < size; x++ {
			switch row[x] {
			case 'C':
				img.Set(x, y, cyan)
			case 'E':
				img.Set(x, y, emerald)
			}
		}
	}

	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		return nil
	}
	pngBytes := pngBuf.Bytes()

	icoBuf := new(bytes.Buffer)
	icoBuf.Write([]byte{0x00, 0x00, 0x01, 0x00, 0x01, 0x00})

	icoBuf.WriteByte(byte(size))
	icoBuf.WriteByte(byte(size))
	icoBuf.WriteByte(0)
	icoBuf.WriteByte(0)
	_ = binary.Write(icoBuf, binary.LittleEndian, uint16(1))
	_ = binary.Write(icoBuf, binary.LittleEndian, uint16(32))
	_ = binary.Write(icoBuf, binary.LittleEndian, uint32(len(pngBytes)))
	_ = binary.Write(icoBuf, binary.LittleEndian, uint32(22))

	icoBuf.Write(pngBytes)
	return icoBuf.Bytes()
}

// ExportDefaultIcon writes default icon to specified path.
func ExportDefaultIcon(path string) error {
	return os.WriteFile(path, GenerateDefaultIcon(), 0644)
}
