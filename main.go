package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/agnivade/levenshtein"
)

const apiKey = "KEY-OPENWEATHER"

type WeatherData struct {
	Name    string `json:"name"`
	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
	} `json:"weather"`
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		city := r.URL.Query().Get("city")
		if city == "" {
			http.Error(w, "Por favor, forneça uma cidade para pesquisar.", http.StatusBadRequest)
			return
		}

		weatherData, err := getWeatherData(city)
		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao obter dados do OpenWeather: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		temperature := int(weatherData.Main.Temp)

		translatedMain := translateToPortuguese(weatherData.Weather[0].Main)
		translatedDescription := translateToPortuguese(weatherData.Weather[0].Description)

		type WeatherResponse struct {
			City        string `json:"city"`
			Temperature int    `json:"temperature"`
			Main        string `json:"main"`
			Description string `json:"description"`
		}

		responseData := WeatherResponse{
			City:        weatherData.Name,
			Temperature: temperature,
			Main:        translatedMain,
			Description: translatedDescription,
		}

		var responseJSON strings.Builder
		encoder := json.NewEncoder(&responseJSON)
		encoder.SetIndent("", "    ")
		err = encoder.Encode(responseData)
		if err != nil {
			http.Error(w, "Erro ao formatar os dados de resposta.", http.StatusInternalServerError)
			return
		}

		cleanedJSON := removeCharsFromJSON(responseJSON.String(), "{}")

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cleanedJSON))
	})

	http.ListenAndServe(":8080", nil)
}

func getWeatherData(city string) (*WeatherData, error) {
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&units=metric&appid=%s", city, apiKey)

	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Não foi possível obter dados da cidade. Código de status: %d", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var data WeatherData
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

func translateToPortuguese(text string) string {
	translations := map[string]string{
		"Clear":        "Limpo",
		"Clouds":       "Nuvens",
		"Rain":         "Chuva",
		"Drizzle":      "Garoa",
		"Thunderstorm": "Trovoadas",
		"Snow":         "Neve",
	}

	lowerText := strings.ToLower(text)
	bestMatch := text
	bestDistance := len(text)

	for key, value := range translations {
		distance := levenshtein.ComputeDistance(lowerText, strings.ToLower(key))
		if distance < bestDistance {
			bestMatch = value
			bestDistance = distance
		}
	}

	return bestMatch
}

func removeCharsFromJSON(jsonData string, charsToRemove string) string {
	removeMap := func(r rune) rune {
		if strings.ContainsRune(charsToRemove, r) {
			return -1
		}
		return r
	}
	return strings.Map(removeMap, jsonData)
}
