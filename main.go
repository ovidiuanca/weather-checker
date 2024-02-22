package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
)

type model struct {
	textInput   textinput.Model
	spinner     spinner.Model
	city        string
	temperature float64
	err         error
	isLoading   bool
}

type temperatureMsg float64
type errMsg error

func getWeather(city string) (float64, error) {
	openWeatherKey := os.Getenv("OPEN_WEATHER_API_KEY")

	if openWeatherKey == "" {
		err := fmt.Errorf("OPEN_WEATHER_API_KEY is not set")

		return -1, err
	}

	city = strings.ReplaceAll(city, " ", "+")
	city = strings.ReplaceAll(city, "-", "%2D")

	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&APPID=%s", city, openWeatherKey)

	c := &http.Client{Timeout: 10 * time.Second}

	res, err := c.Get(url)

	if err != nil {
		return 0, err
	}

	defer res.Body.Close() // nolint:errcheck

	var data map[string]any

	json.NewDecoder(res.Body).Decode(&data)

	var temperature float64

	switch data["cod"].(type) {
	case float64:
		switch data["cod"].(float64) {
		case 200:
			tempKelvin := data["main"].(map[string]any)["temp"].(float64)

			temperature = tempKelvin - 273.15
		case 401:
			temperature = -1
			err = fmt.Errorf("invalid API key")
		default:
			temperature = -1
			err = fmt.Errorf("something went wrong")
		}
	default:
		temperature = -1
		err = fmt.Errorf("city \"%s\" not found", city)
	}

	return temperature, err
}

func InitialModel() model {
	textInputModel := textinput.New()

	textInputModel.Placeholder = "Frankfurt"
	textInputModel.Width = 20
	textInputModel.CharLimit = 50
	textInputModel.Focus()

	spinnerModel := spinner.New()
	spinnerType := spinner.Dot
	spinnerModel.Spinner = spinnerType

	return model{
		textInput: textInputModel,
		spinner:   spinnerModel,
		err:       nil,
		isLoading: false,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.city = m.textInput.Value()
			m.isLoading = true

			return m, func() tea.Msg {
				temp, err := getWeather(m.city)

				if err != nil {
					return errMsg(err)
				}

				return temperatureMsg(temp)
			}

		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case temperatureMsg:
		m.temperature = float64(msg)
		m.isLoading = false
		return m, nil
	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		m.isLoading = false
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf(
			"\nError: %s\n\n%s", m.err,
			"(esc to quit)",
		)
	} else {
		if m.isLoading {
			return fmt.Sprintf(
				"\n%s Loading...\n\n%s",
				m.spinner.View(),
				"(esc to quit)",
			)
		} else {
			if m.temperature == 0 {
				return fmt.Sprintf(
					"\nEnter a city end press Enter?\n\n%s\n\n%s",
					m.textInput.View(),
					"(esc to quit)",
				) + "\n"
			} else {
				return fmt.Sprintf(
					"\nTemperature in %s is %.1f°C\n\n%s",
					m.city,
					m.temperature,
					"(esc to quit)",
				) + "\n"
			}
		}
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
