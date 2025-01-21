package draw

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/gui/assets/icon"
	"github.com/g3n/engine/math32"
)

type FileSelect struct {
	gui.Panel
	path         *gui.Label
	list         *gui.List
	bok          *gui.Button
	bcan         *gui.Button
	onFileSelect func(string) error
}

func NewFileSelect(width, height float32, onFileSelect func(string) error) (*FileSelect, error) {
	fs := new(FileSelect)
	fs.Panel.Initialize(fs, width, height)
	fs.SetBorders(2, 2, 2, 2)
	fs.SetPaddings(4, 4, 4, 4)
	fs.SetColor(math32.NewColor("White"))
	fs.SetVisible(false)
	fs.SetBounded(false)

	// Imposta il layout verticale per l'intero pannello
	l := gui.NewVBoxLayout()
	l.SetSpacing(4)
	fs.SetLayout(l)

	// Crea l'etichetta del percorso
	fs.path = gui.NewLabel("path")
	fs.Add(fs.path)

	// Crea la lista
	fs.list = gui.NewVList(0, 0)
	fs.list.SetLayoutParams(&gui.VBoxLayoutParams{Expand: 5, AlignH: gui.AlignWidth})
	fs.list.Subscribe(gui.OnChange, func(evname string, ev interface{}) {
		fs.handleSelect()
	})
	fs.Add(fs.list)

	// Pannello contenitore per i pulsanti
	bc := gui.NewPanel(0, 0)
	bcl := gui.NewHBoxLayout()
	bcl.SetAlignH(gui.AlignWidth)
	bc.SetLayout(bcl)
	bc.SetLayoutParams(&gui.VBoxLayoutParams{Expand: 1, AlignH: gui.AlignWidth})
	fs.Add(bc)

	// Crea il pulsante OK
	fs.bok = gui.NewButton("OK")
	fs.bok.SetLayoutParams(&gui.HBoxLayoutParams{Expand: 0, AlignV: gui.AlignCenter})
	fs.onFileSelect = onFileSelect
	fs.bok.Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		selected := fs.Selected()
		if selected != "" && filepath.Ext(selected) == ".fss" {
			if err := fs.onFileSelect(selected); err != nil {
				// TODO: Show error to user
				return
			}
			fs.Show(false)
		}
	})
	bc.Add(fs.bok)

	// Crea il pulsante Annulla
	fs.bcan = gui.NewButton("Cancel")
	fs.bcan.SetLayoutParams(&gui.HBoxLayoutParams{Expand: 0, AlignV: gui.AlignCenter})
	fs.bcan.Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		fs.Dispatch("OnCancel", nil)
	})
	bc.Add(fs.bcan)

	// Imposta la directory iniziale
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	} else {
		fs.SetPath(path)
	}
	return fs, nil
}

// Mostra o nasconde la finestra di dialogo per la selezione dei file
func (fs *FileSelect) Show(show bool) {
	if show {
		fs.SetVisible(true)
		width, height := app.App().GetSize()
		px := (float32(width) - fs.Width()) / 2
		py := (float32(height) - fs.Height()) / 2
		fs.SetPosition(px, py)
	} else {
		fs.SetVisible(false)
	}
}

func (fs *FileSelect) SetPath(path string) error {
	// Apre il file o la directory
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Verifica se è una directory
	files, err := f.Readdir(0)
	if err != nil {
		return err
	}
	fs.path.SetText(path)

	// Ordina i file per nome
	sort.Sort(listFileInfo(files))

	// Legge il contenuto della directory e lo carica nella lista
	fs.list.Clear()
	// Aggiunge la directory precedente
	prev := gui.NewImageLabel("..")
	prev.SetIcon(icon.FolderOpen)
	fs.list.Add(prev)
	// Aggiunge i file della directory
	for i := 0; i < len(files); i++ {
		item := gui.NewImageLabel(files[i].Name())
		if files[i].IsDir() {
			item.SetIcon(icon.FolderOpen)
		} else if filepath.Ext(files[i].Name()) == ".fss" {
			item.SetIcon(icon.Description)
		} else {
			continue // Skip non-FSS files
		}
		fs.list.Add(item)
	}
	return nil
}

func (fs *FileSelect) Selected() string {
	selist := fs.list.Selected()
	if len(selist) == 0 {
		return ""
	}
	label := selist[0].(*gui.ImageLabel)
	text := label.Text()
	return filepath.Join(fs.path.Text(), text)
}

func (fs *FileSelect) handleSelect() {
	// Ottiene l'etichetta dell'immagine selezionata e il suo testo
	sel := fs.list.Selected()[0]
	label := sel.(*gui.ImageLabel)
	text := label.Text()

	// Verifica se è la directory precedente
	if text == ".." {
		dir, _ := filepath.Split(fs.path.Text())
		fs.SetPath(filepath.Dir(dir))
		return
	}

	// Verifica se è una directory
	path := filepath.Join(fs.path.Text(), text)
	s, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	if s.IsDir() {
		fs.SetPath(path)
	}
}

// Per ordinare l'array di FileInfo per Nome
type listFileInfo []os.FileInfo

func (fi listFileInfo) Len() int      { return len(fi) }
func (fi listFileInfo) Swap(i, j int) { fi[i], fi[j] = fi[j], fi[i] }
func (fi listFileInfo) Less(i, j int) bool {
	return fi[i].Name() < fi[j].Name()
}
