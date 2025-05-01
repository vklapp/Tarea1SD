package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const baseURL = "http://10.10.28.60:8080/api"
func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println(`
Menu
1. Ver corredores
2. Ver detalle de corredor
3. Ver carreras
4. Ver detalle de carrera
5. Resumen de temporada
6. Salir`)
		fmt.Print("Seleccione una opción: ")
		scanner.Scan()
		opcion := scanner.Text()

		switch opcion {
		case "1":
			verCorredores()
		case "2":
			fmt.Print("Ingrese el número del piloto: ")
			scanner.Scan()
			num := scanner.Text()
			verDetalleCorredor(num)
		case "3":
			verCarreras()
		case "4":
			fmt.Print("Ingrese el ID de la carrera: ")
			scanner.Scan()
			num := scanner.Text()
			verDetalleCarrera(num)
		case "5":
			verResumenTemporada()
		case "6":
			fmt.Println("Fin del programa.")
			return
		default:
			fmt.Println("Opción inválida.")
		}
	}
}

func verCorredores() {
	resp, err := http.Get(baseURL + "/corredor")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	var pilotos []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&pilotos)

	fmt.Println("-------------------------------------------------------------")
	fmt.Println("| # | Nombre | Apellido | Nº Piloto | Equipo | País |")
	fmt.Println("-------------------------------------------------------------")
	for i, p := range pilotos {
		fmt.Printf("| %d | %s | %s | %v | %s | %s |\n",
			i+1, p["first_name"], p["last_name"], p["driver_number"], p["team_name"], p["country_code"])
	}
	fmt.Println("-------------------------------------------------------------")
}

func verDetalleCorredor(id string) {
	resp, err := http.Get(baseURL + "/corredor/detalle/" + id)
	if err != nil {
		fmt.Println("Error al conectar con el servidor:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("Error del servidor:", resp.Status)
		return
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Println("Error al decodificar la respuesta:", err)
		return
	}

	val, ok := data["race_results"]
	if !ok || val == nil {
		fmt.Println("No hay resultados disponibles para este piloto.")
		return
	}
	resultados := val.([]interface{})

	fmt.Println("-------------------------------------------------------------------------------------------")
	fmt.Println("| # | Carrera | Pos Final | Vuelta rápida | Velocidad máx | Menor tiempo vuelta |")
	fmt.Println("-------------------------------------------------------------------------------------------")
	for i, r := range resultados {
		row := r.(map[string]interface{})
		fmt.Printf("| %d | %s | %v | %v | %.0f km/h | %.3f s |\n",
			i+1,
			row["race"],
			row["position"],
			boolToStr(row["fastest_lap"].(bool)),
			row["max_speed"].(float64),
			row["best_lap_duration"].(float64))
	}
	fmt.Println("-------------------------------------------------------------------------------------------")

	summaryVal, ok := data["performance_summary"]
	if !ok || summaryVal == nil {
		fmt.Println("No hay resumen disponible.")
		return
	}
	summary := summaryVal.(map[string]interface{})

	fmt.Println("-----------------------------------------------")
	fmt.Println("| Resumen del desempeño del piloto           |")
	fmt.Println("-----------------------------------------------")
	fmt.Printf("| Carreras ganadas:           | %v |\n", summary["wins"])
	fmt.Printf("| Veces en el top 3:          | %v |\n", summary["top_3_finishes"])
	fmt.Printf("| Velocidad máxima alcanzada: | %.0f km/h |\n", summary["max_speed"].(float64))
	fmt.Println("-----------------------------------------------")
}


func verCarreras() {
	resp, err := http.Get(baseURL + "/carrera")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	var carreras []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&carreras)

	fmt.Println("-------------------------------------------------------------------------")
	fmt.Println("| # | ID carrera | País | Fecha | Año | Circuito |")
	fmt.Println("-------------------------------------------------------------------------")
	for i, c := range carreras {
		fmt.Printf("| %d | %v | %s | %s | %v | %s |\n",
			i+1, c["session_key"], c["country_name"], formatFecha(c["date_start"].(string)), c["year"], c["circuit_short_name"])
	}
	fmt.Println("-------------------------------------------------------------------------")
}

func verDetalleCarrera(id string) {
	resp, err := http.Get(baseURL + "/carrera/detalle/" + id)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	fmt.Println("---------------------------------------------------------------")
	fmt.Println("| Resultados |")
	fmt.Println("---------------------------------------------------------------")
	fmt.Println("| Posición | Piloto | Equipo | País |")
	fmt.Println("---------------------------------------------------------------")
	for _, r := range data["results"].([]interface{}) {
		row := r.(map[string]interface{})
		fmt.Printf("| %v | %s | %s | %s |\n", row["position"], row["driver"], row["team"], row["country"])
	}
	fmt.Println("---------------------------------------------------------------")

	fmt.Println("| Vuelta más rápida |")
	vl := data["fastest_lap"].(map[string]interface{})
	fmt.Println("---------------------------------------------------------------")
	fmt.Printf("| Piloto: %s | Total: %.3f s | Sectores: %.3f / %.3f / %.3f |\n",
		vl["driver"], vl["total_time"], vl["sector_1"], vl["sector_2"], vl["sector_3"])
	fmt.Println("---------------------------------------------------------------")

	fmt.Println("| Velocidad máxima |")
	vs := data["max_speed"].(map[string]interface{})
	fmt.Printf("| Piloto: %s | Velocidad: %.1f km/h |\n", vs["driver"], vs["speed_kmh"])
	fmt.Println("---------------------------------------------------------------")
}

func verResumenTemporada() {
	resp, err := http.Get(baseURL + "/temporada/resumen")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	var resumen map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&resumen)

	Resumen("Victorias", "top_3_winners", resumen)
	Resumen("Vueltas Rapidas", "top_3_fastest_laps", resumen)
	Resumen("Pole Positions", "top_3_pole_positions", resumen)
}

func Resumen(titulo, clave string, resumen map[string]interface{}) {
	lista := resumen[clave].([]interface{})
	header := fmt.Sprintf("| Top 3 Pilotos con mas %s - Temporada 2024 |", titulo)
	border := strings.Repeat("-", len(header)-2)

	fmt.Println()
	fmt.Println(header)
	fmt.Printf("|%s|\n", border)
	fmt.Printf("| %-8s | %-16s | %-10s | %-4s | %-15s |\n", "Posicion", "Piloto", "Equipo", "Pais", titulo)
	fmt.Printf("|%s|\n", border)

	for _, r := range lista {
		fila := r.(map[string]interface{})
		pos := int(fila["position"].(float64))
		nombre := fila["driver"].(string)
		equipo := fila["team_name"].(string)
		pais := fila["country_code"].(string)
		val := int(fila["Value"].(float64))

		fmt.Printf("| %-8d | %-16s | %-10s | %-4s | %-15d |\n", pos, nombre, equipo, pais, val)
	}
	fmt.Printf("|%s|\n", border)
}

func printTop(lista interface{}) {
	for _, r := range lista.([]interface{}) {
		row := r.(map[string]interface{})
		val := row["Value"]
		if val == nil {
			val = row["wins"]
			if val == nil {
				val = row["fastest_laps"]
			}
			if val == nil {
				val = row["poles"]
			}
		}
		fmt.Printf("| %v | %s | %v |\n", row["position"], row["driver"], val)
	}
}

func boolToStr(b bool) string {
	if b {
		return "Sí"
	}
	return "No"
}

func formatFecha(s string) string {
	// Quita la T y lo que venga después
	if i := strings.Index(s, "T"); i != -1 {
		return s[:i]
	}
	return s
}
