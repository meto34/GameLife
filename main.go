package main

import (
	"errors"
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const cellSize = 40

type Cell struct {
	widget.BaseWidget
	alive    *bool
	rect     *canvas.Rectangle
	editable bool
}

func hasAliveCells(field [][]bool) bool {

	for i := range field {
		for j := range field[i] {

			if field[i][j] {
				return true
			}
		}
	}

	return false
}
func equalFields(a, b [][]bool) bool {

	if len(a) != len(b) {
		return false
	}

	if len(a[0]) != len(b[0]) {
		return false
	}

	for i := range a {
		for j := range a[i] {

			if a[i][j] != b[i][j] {
				return false
			}
		}
	}

	return true
}

func NewCell(alive *bool, editable bool) *Cell {

	c := &Cell{
		editable: editable,
		alive:    alive,
	}
	if *alive {
		c.rect = canvas.NewRectangle(color.RGBA{0, 255, 0, 255})
	} else {
		c.rect = canvas.NewRectangle(color.RGBA{255, 0, 0, 255})
	}
	c.ExtendBaseWidget(c)
	return c
}
func (c *Cell) Tapped(*fyne.PointEvent) {
	if !c.editable {
		return
	}
	*c.alive = !*c.alive

	if *c.alive {
		c.rect.FillColor = color.RGBA{0, 255, 0, 255}
	} else {
		c.rect.FillColor = color.RGBA{255, 0, 0, 255}
	}

	c.rect.Refresh()
}

func (c *Cell) TappedSecondary(*fyne.PointEvent) {}

func (c *Cell) CreateRenderer() fyne.WidgetRenderer {
	c.rect.SetMinSize(fyne.NewSize(cellSize, cellSize))

	return widget.NewSimpleRenderer(c.rect)
}

func countNeighbors(field [][]bool, row, col int) int {

	rows := len(field)
	cols := len(field[0])

	count := 0

	for dr := -1; dr <= 1; dr++ {
		for dc := -1; dc <= 1; dc++ {

			// пропускаем саму клетку
			if dr == 0 && dc == 0 {
				continue
			}

			r := row + dr
			c := col + dc

			// проверка границ
			if r >= 0 && r < rows &&
				c >= 0 && c < cols {

				if field[r][c] {
					count++
				}
			}
		}
	}

	return count
}

func nextGeneration(field [][]bool) [][]bool {

	rows := len(field)
	cols := len(field[0])

	next := make([][]bool, rows)

	for i := range next {
		next[i] = make([]bool, cols)
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {

			neighbors := countNeighbors(field, i, j)

			// если клетка живая
			if field[i][j] {

				// живет при 2 или 3 соседях
				next[i][j] = neighbors == 2 || neighbors == 3

			} else {

				// мертвая оживает при 3 соседях
				next[i][j] = neighbors == 3
			}
		}
	}

	return next
}

func buildGrid(field [][]bool, editable bool) *fyne.Container {

	rows := len(field)
	cols := len(field[0])

	grid := container.NewGridWithColumns(cols)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {

			cell := NewCell(&field[i][j], editable)

			grid.Add(cell)
		}
	}

	return grid
}

// с помощью этой функции сохраняю значения поколений
func copyField(field [][]bool) [][]bool {

	rows := len(field)

	copyArr := make([][]bool, rows)

	for i := range field {

		copyArr[i] = make([]bool, len(field[i]))

		copy(copyArr[i], field[i])
	}

	return copyArr
}
func checkGameOver(
	currentField [][]bool,
	nextField [][]bool,
	history [][][]bool,
) (bool, string) {

	// 1. Нет живых клеток
	if !hasAliveCells(nextField) {
		return true, "на поле не осталось живых клеток"
	}

	// 2. Стабильная конфигурация
	if equalFields(currentField, nextField) {
		return true, "достигнута стабильная конфигурация"
	}

	// 3. Периодическая конфигурация
	for _, oldField := range history {

		if equalFields(oldField, nextField) {
			return true, "обнаружена периодическая конфигурация"
		}
	}

	return false, ""
}

func createField(
	win fyne.Window,
	inputVert *widget.Entry,
	inputHor *widget.Entry,
) func() {

	return func() {

		// Читаем вертикальный размер
		sizeVert, err := strconv.Atoi(inputVert.Text)
		if err != nil {
			dialog.ShowError(
				errors.New("вертикальный размер должен быть числом"),
				win,
			)
			return
		}

		if sizeVert <= 0 {
			dialog.ShowError(
				errors.New("вертикальный размер должен быть больше 0"),
				win,
			)
			return
		}

		// Читаем горизонтальный размер
		sizeHor, err := strconv.Atoi(inputHor.Text)
		if err != nil {
			dialog.ShowError(
				errors.New("горизонтальный размер должен быть числом"),
				win,
			)
			return
		}

		if sizeHor <= 0 {
			dialog.ShowError(
				errors.New("горизонтальный размер должен быть больше 0"),
				win,
			)
			return
		}

		field := make([][]bool, sizeVert)

		for i := range field {
			field[i] = make([]bool, sizeHor)

			for j := range field[i] {
				field[i][j] = true // все клетки живые
			}
		}
		currentField := field
		history := make([][][]bool, 0)
		history = append(history, copyField(currentField))
		var nextField [][]bool

		currentGrid := buildGrid(currentField, true)

		var nextGrid *fyne.Container

		started := false

		var startBtn *widget.Button

		startBtn = widget.NewButton("Начать игру", func() {

			// Здесь поведение при первом запуске
			if !started {
				currentGrid = buildGrid(currentField, false)
				nextField = nextGeneration(currentField)

				gameOver, reason := checkGameOver(
					currentField,
					nextField,
					history,
				)

				if gameOver {

					dialog.ShowInformation(
						"Игра окончена",
						"Игра окончена из-за того, что "+reason,
						win,
					)

					startBtn.Disable()
				}

				history = append(history, copyField(nextField))

				nextGrid = buildGrid(nextField, false)

				started = true

				startBtn.SetText("Следующий шаг")

				win.SetContent(container.NewVBox(
					startBtn,

					widget.NewLabel("Текущее поколение"),
					currentGrid,

					widget.NewLabel("Следующее поколение"),
					nextGrid,
				))

				return
			}

			currentField = nextField

			nextField = nextGeneration(currentField)

			gameOver, reason := checkGameOver(
				currentField,
				nextField,
				history,
			)

			if gameOver {

				dialog.ShowInformation(
					"Игра окончена",
					"Игра окончена из-за того, что "+reason,
					win,
				)

				startBtn.Disable()
			}

			history = append(history, copyField(nextField))

			currentGrid = buildGrid(currentField, false)
			nextGrid = buildGrid(nextField, false)

			win.SetContent(container.NewVBox(
				startBtn,

				widget.NewLabel("Текущее поколение"),
				currentGrid,

				widget.NewLabel("Следующее поколение"),
				nextGrid,
			))
		})
		win.SetContent(container.NewVBox(
			startBtn,

			widget.NewLabel("Стартовое поколение"),
			currentGrid,
		))
	}

}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Жизнь")

	labelInputVert := widget.NewLabel("Введите размер поля по вертикали:")
	inputVert := widget.NewEntry()

	labelInputHor := widget.NewLabel("Введите размер поля по горизонтали:")
	inputHor := widget.NewEntry()

	okbtn := widget.NewButton(
		"Подтвердить",
		createField(myWindow, inputVert, inputHor),
	)

	content := container.NewVBox(
		labelInputVert,
		inputVert,
		labelInputHor,
		inputHor,
		okbtn,
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(800, 800))
	myWindow.ShowAndRun()
}
