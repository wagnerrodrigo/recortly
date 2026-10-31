package template

// Preset é um modelo pronto mostrado na tela "Escolha um template".
type Preset struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Spec        Spec   `json:"spec"`
}

func Presets() []Preset {
	twitter := Default()
	twitter.Name = "Twitter / X"
	twitter.Preset = "twitter"
	twitter.Canvas.Background = "#FFFFFF"
	twitter.Text.Color = "#0F1419"
	twitter.Text.Band = false
	twitter.Text.StrokeEnabled = false
	twitter.Text.Position = "top"
	twitter.Text.Size = 46
	twitter.Profile.Position = "top"
	twitter.Profile.Align = "left"
	twitter.Profile.Theme = "light"
	twitter.Profile.NameSize = 48
	twitter.Profile.HandleSize = 40
	twitter.Zones = []Zone{{Type: "video", Weight: 1, Radius: 20, Fit: "cover"}}
	twitter.Phrases = []string{"Olha só o que rolou nesse servidor"}

	meme := Default()
	meme.Name = "Meme"
	meme.Preset = "meme"
	meme.Canvas.Background = "#000000"
	meme.Text.Color = "#FFFFFF"
	meme.Text.Band = false
	meme.Text.StrokeEnabled = true
	meme.Text.StrokeWidth = 5
	meme.Text.StrokeColor = "#000000"
	meme.Text.Position = "top"
	meme.Text.Size = 54
	meme.Profile.Position = "bottom"
	meme.Profile.Align = "center"
	meme.Profile.Theme = "dark"
	meme.Zones = []Zone{{Type: "video", Weight: 1, Radius: 0, Fit: "contain"}}
	meme.Phrases = []string{"POV: você entra no lobby errado"}

	viral := Default()
	viral.Name = "Vídeo viral"
	viral.Preset = "viral"
	viral.Canvas.Background = "#0E0E10"
	viral.Text.Color = "#FFFFFF"
	viral.Text.Band = true
	viral.Text.BandColor = "#000000"
	viral.Text.StrokeEnabled = false
	viral.Text.Position = "bottom"
	viral.Text.Size = 56
	viral.Profile.Position = "top"
	viral.Profile.Align = "center"
	viral.Profile.Theme = "dark"
	viral.Zones = []Zone{
		{Type: "video", Weight: 0.62, Radius: 18, Fit: "cover"},
		{Type: "blank", Weight: 0.38, Radius: 0, Fit: "cover"},
	}
	viral.Phrases = []string{"Isso aqui viralizou e você ainda não sabia"}

	blank := Default()
	blank.Name = "Upload sem template"
	blank.Preset = "blank"
	blank.Profile.Position = "hidden"
	blank.Text.Position = "hidden"
	blank.Text.Band = false
	blank.Zones = []Zone{{Type: "video", Weight: 1, Radius: 0, Fit: "cover"}}
	blank.Phrases = []string{" "}

	return []Preset{
		{ID: "twitter", Name: twitter.Name, Description: "Card branco com perfil, headline e vídeo vertical.", Spec: twitter},
		{ID: "meme", Name: meme.Name, Description: "Texto grande com contorno, vídeo central e perfil embaixo.", Spec: meme},
		{ID: "viral", Name: viral.Name, Description: "Headline em faixa, vídeo no topo e espaço para respiro.", Spec: viral},
		{ID: "blank", Name: blank.Name, Description: "Só o vídeo em 9:16, sem perfil nem texto.", Spec: blank},
	}
}

func PresetByID(id string) (Preset, bool) {
	for _, p := range Presets() {
		if p.ID == id {
			return p, true
		}
	}
	return Preset{}, false
}
