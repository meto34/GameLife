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
type Game struct {
	win fyne.Window

	startBtn *widget.Button

	currentField [][]bool
	nextField    [][]bool

	history [][][]bool

	started bool
}

func (g *Game) render() {

	currentGrid := buildGrid(g.currentField, false)
	nextGrid := buildGrid(g.nextField, false)

	g.win.SetContent(container.NewVBox(
		g.startBtn,

		widget.NewLabel("Текущее поколение"),
		currentGrid,

		widget.NewLabel("Следующее поколение"),
		nextGrid,
	))
}
func (g *Game) checkGameOver() bool {
	return g.handleGameOver()
}
func (g *Game) firstStep() {

	g.nextField = nextGeneration(g.currentField)

	if g.checkGameOver() {
		return
	}

	g.history = append(
		g.history,
		copyField(g.nextField),
	)

	g.started = true

	g.startBtn.SetText("Следующий шаг")

	g.render()
}
func (g *Game) nextStep() {

	g.currentField = g.nextField

	g.nextField = nextGeneration(g.currentField)

	if g.checkGameOver() {
		return
	}

	g.history = append(
		g.history,
		copyField(g.nextField),
	)

	g.render()
}
func (g *Game) createGameHandler() func() {

	return func() {

		if !g.started {
			g.firstStep()
			return
		}

		g.nextStep()
	}
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
func survive(alive bool, neighbors int) bool {
	if alive {
		return neighbors == 2 || neighbors == 3
	}
	return neighbors == 3
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
			next[i][j] = survive(field[i][j], neighbors)
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
func isGameOver(
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
func (g *Game) handleGameOver() bool {

	gameOver, reason := isGameOver(
		g.currentField,
		g.nextField,
		g.history,
	)

	if !gameOver {
		return false
	}

	dialog.ShowInformation(
		"Игра окончена",
		"Игра окончена из-за того, что "+reason,
		g.win,
	)

	g.startBtn.Disable()

	return true
}

func (g *Game) step() bool {

	if !g.started {
		g.nextField = nextGeneration(g.currentField)
		g.started = true
		g.startBtn.SetText("Следующий шаг")
		return true
	}

	g.currentField = g.nextField
	g.nextField = nextGeneration(g.currentField)

	return true
}
func (g *Game) handler() func() {

	return func() {

		g.step()

		if g.checkGameOver() {
			return
		}

		g.history = append(
			g.history,
			copyField(g.nextField),
		)

		g.render()
	}
}
func createField(
	win fyne.Window,
	inputVert *widget.Entry,
	inputHor *widget.Entry,
) func() {
	return func() {
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
		game := &Game{
			win: win,
		}
		for i := range field {
			field[i] = make([]bool, sizeHor)

			for j := range field[i] {
				field[i][j] = true // все клетки живые по умолчанию
			}
		}
		game.currentField = field
		game.history = append(game.history, copyField(field))

		currentGrid := buildGrid(game.currentField, true)

		game.startBtn = widget.NewButton("Начать игру", game.handler())

		win.SetContent(container.NewVBox(
			game.startBtn,

			widget.NewLabel("Стартовое поколение"),
			currentGrid,
		))
	}

}
func makeStartScreen(win fyne.Window) fyne.CanvasObject {

	labelInputVert := widget.NewLabel("Введите размер поля по вертикали:")
	inputVert := widget.NewEntry()
	inputVert.SetPlaceHolder("5")

	labelInputHor := widget.NewLabel("Введите размер поля по горизонтали:")
	inputHor := widget.NewEntry()
	inputHor.SetPlaceHolder("4")

	okbtn := widget.NewButton(
		"Подтвердить",
		createField(win, inputVert, inputHor),
	)

	return container.NewVBox(
		labelInputVert,
		inputVert,
		labelInputHor,
		inputHor,
		okbtn,
	)
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Жизнь")

	myWindow.SetContent(makeStartScreen(myWindow))
	myWindow.Resize(fyne.NewSize(800, 800))
	myWindow.ShowAndRun()
}
