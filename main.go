package main

import (
	"image"
	"log"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	giotext "gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	qrcode "github.com/skip2/go-qrcode"
	"golang.design/x/clipboard"
)

func main() {
	// Инициализация доступа к буферу обмена (обязательно для Linux)
	err := clipboard.Init()
	if err != nil {
		log.Fatal(err)
	}

	// Запускаем окно в отдельной горутине (требование Gio)
	go func() {
		w := new(app.Window)
		w.Option(
			app.Title("QR Code from Clipboard"),
			app.Size(unit.Dp(500), unit.Dp(500)),
		)
		if err := run(w); err != nil {
			log.Fatal(err)
		}
	}()
	app.Main() // запускает главный цикл событий
}

func run(w *app.Window) error {
	// Читаем текст из буфера обмена
	text := clipboard.Read(clipboard.FmtText)
	if len(text) == 0 {
		text = []byte("Clipboard is empty")
	}

	// Генерируем QR‑код (размер 256×256 пикселей)
	qr, err := qrcode.New(string(text), qrcode.Medium)
	if err != nil {
		qr, _ = qrcode.New("QR generation error", qrcode.Medium)
	}
	img := qr.Image(256) // *image.RGBA

	// Подготавливаем виджет для отображения изображения
	imgOp := paint.NewImageOp(img)
	var imgWidget widget.Image
	imgWidget.Src = imgOp
	imgWidget.Fit = widget.ScaleDown // масштабирование с сохранением пропорций

	var ops op.Ops
	th := material.NewTheme()
	for {
		e := w.Event()
		switch e := e.(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			handleFrameEvent(&ops, e, &imgWidget, th, string(text))
		}
	}
}

func handleFrameEvent(ops *op.Ops, e app.FrameEvent, imgWidget *widget.Image, th *material.Theme, clipText string) {
	gtx := app.NewContext(ops, e)

	// Оборачиваем всё в layout.Center для вертикального центрирования
	layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Вертикальный макет для QR-кода и текста
		return layout.Flex{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,      // выравнивание по вертикали внутри Flex
			Spacing:   layout.SpaceEvenly, // равномерное распределение пространства
		}.Layout(gtx,
			// 1. QR-код
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				maxSize := gtx.Constraints.Max.X
				if gtx.Constraints.Max.Y < maxSize {
					maxSize = gtx.Constraints.Max.Y
				}
				margin := 40
				maxSize -= margin * 2
				if maxSize < 10 {
					maxSize = 10
				}

				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Max = image.Point{X: maxSize, Y: maxSize}
					return imgWidget.Layout(gtx)
				})
			}),

			// 2. Текст под QR-кодом
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Label(th, unit.Sp(16), clipText)
				label.Alignment = giotext.Middle
				label.MaxLines = 3

				gtx.Constraints.Max.X -= 40
				gtx.Constraints.Min.X = 0

				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return label.Layout(gtx)
				})
			}),
		)
	})

	e.Frame(gtx.Ops)
}
