package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
)

type model struct {
	textInput   textinput.Model
	city        string
	temperature float64
	err         error
}

type temperatureMsg float64
type errMsg error

func getWeather(city string) (float64, error) {
	openWeatherKey := os.Getenv("OPEN_WEATHER_API_KEY")

	if openWeatherKey == "" {
		panic("OPEN_WEATHER_API_KEY is not set")
	}

	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&APPID=%s", city, openWeatherKey)

	c := &http.Client{Timeout: 10 * time.Second}

	res, err := c.Get(url)

	if err != nil {
		return 0, err
	}

	defer res.Body.Close() // nolint:errcheck

	var data map[string]interface{}

	json.NewDecoder(res.Body).Decode(&data)

	mainData := data["main"]

	var temperature float64

	if mainData != nil {
		tempKelvin := mainData.(map[string]interface{})["temp"].(float64)

		temperature = tempKelvin - 273.15
	} else {
		err = fmt.Errorf("city not found, do not type spaces, special characters or numbers")
	}

	return temperature, err
}

func InitialModel() model {
	textInputModel := textinput.New()

	textInputModel.Placeholder = "Frankfurt"
	textInputModel.Width = 20
	textInputModel.CharLimit = 50
	textInputModel.Focus()

	return model{
		textInput: textInputModel,
		err:       nil,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyEnter:
			m.city = m.textInput.Value()

			return m, func() tea.Msg {
				temp, _ := getWeather(m.city)

				return temperatureMsg(temp)
			}

		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	case temperatureMsg:
		m.temperature = float64(msg)
		return m, nil
	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.temperature == 0 {
		return fmt.Sprintf(
			"Enter a city end press Enter?\n\n%s\n\n%s",
			m.textInput.View(),
			"(esc to quit)",
		) + "\n"
	} else if m.err != nil {
		return fmt.Sprintf("Something went wrong: %s", m.err)
	} else {
		return fmt.Sprintf(
			"Temperature in %s is %.1f°C\n\n%s",
			m.city,
			m.temperature,
			"(esc to quit)",
		) + "\n"
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	if _, err := tea.NewProgram(InitialModel()).Run(); err != nil {
		os.Exit(1)
	}
}
