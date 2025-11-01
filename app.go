package main

import (
    "fmt"
    "net/http"
    "os"
    "log/slog"
    "encoding/json"
    "bufio"
    "strings"
)

var PICTURE_PATH string;
var APP_PORT string;

func init() {
    slog.Info("Loading...")
    err := loadEnv(".env")
    if err != nil {
        slog.Error("could not read .env file")
    }

    PICTURE_PATH = os.Getenv("PICTURE_PATH")
    APP_PORT = os.Getenv("APP_PORT")
    if APP_PORT == ""{
        APP_PORT = "8080"
    }
}

func loadEnv(file string) error {
    f, err := os.Open(file)
    if err != nil {
        slog.Error("could not load .env file")
        return err
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || strings.HasPrefix(line, "#") {
            continue // skip empty lines or comments
        }

        parts := strings.SplitN(line, "=", 2)
        if len(parts) != 2 {
            continue
        }

        key := strings.TrimSpace(parts[0])
        value := strings.TrimSpace(parts[1])
        os.Setenv(key, value)
    }

    return scanner.Err()
}

func getDirectoriesName(path string, wantDir bool) []string {
    entries, err := os.ReadDir(path)
    if err != nil {
        return []string{}
    }

    var dirs []string
    for _, e := range entries {

        if wantDir {
            if e.IsDir() {
                dirs = append(dirs, e.Name())
            }
        }else {
            if !e.IsDir() {
                dirs = append(dirs, e.Name())
            }
        }
        
    }

    return dirs
}

func getAll(w http.ResponseWriter, req *http.Request) {

    dirs := getDirectoriesName(PICTURE_PATH,true)

    response := map[string][]string{
        "entity_dir": dirs,
    }

    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func getImageNameList(w http.ResponseWriter, req *http.Request) {

    w.Header().Set("Access-Control-Allow-Origin", "*")
    name := req.URL.Query().Get("name")
    if name == "" {
        http.Error(w, "Missing 'name' parameter", http.StatusBadRequest)
        return
    }

    files := getDirectoriesName(PICTURE_PATH+name,false)

    httpPaths := []string{}
    for _, file := range files {
            // Return relative HTTP path for the image
            httpPaths = append(httpPaths, fmt.Sprintf("images/%s/%s", name, file))
    }

    response := map[string][]string{
        "entity_dir": httpPaths,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func main() {

    http.HandleFunc("/getall", getAll)
    http.HandleFunc("/imagenames", getImageNameList)

    http.Handle("/", http.FileServer(http.Dir("./frontend")))
    http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(PICTURE_PATH))))

    addr := fmt.Sprintf(":%s", APP_PORT)
    slog.Info("server started on", "port", APP_PORT)
    http.ListenAndServe(addr, nil)
}
