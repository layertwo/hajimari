package handlers

import (
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/toboshii/hajimari/internal/config"
	"github.com/toboshii/hajimari/internal/hajimari"
	"github.com/toboshii/hajimari/internal/hajimari/customapps"
	"github.com/toboshii/hajimari/internal/hajimari/ingressapps"
	"github.com/toboshii/hajimari/internal/kube"
	"github.com/toboshii/hajimari/internal/kube/util"
)

type Page struct {
	conf *config.Config
	tpl  *template.Template
}

func NewPage(conf *config.Config, tpl *template.Template) *Page {
	// for i, module := range conf.Modules {
	// 	switch module.Name {
	// 	case "Weather":
	// 		dispatch(&conf.Modules[i], modules.GetWeather)
	// 	case "Pi-hole":
	// 		dispatch(&conf.Modules[i], modules.GetPiholeStats)
	// 	}
	// }

	return &Page{conf, tpl}
}

// func dispatch(module *config.Module, f func(map[string]string) string) {
// 	ticker := time.NewTicker(1 * time.Second)
// 	go func() {
// 		for {
// 			<-ticker.C
// 			ch := make(chan string)
// 			go func() {
// 				ticker = time.NewTicker(module.UpdateInterval * time.Minute)
// 				weather := f(module.Data)
// 				ch <- weather
// 			}()

// 			module.Output = <-ch
// 		}
// 	}()
// }

func (p *Page) StartpageHandler(w http.ResponseWriter, r *http.Request) {
	kubeClient := kube.GetClient()

	var hajimariApps []hajimari.App

	appConfig, err := config.GetConfig()
	if err != nil {
		logger.Error("Failed to read configuration for hajimari", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ingressAppsList := ingressapps.NewList(kubeClient, *appConfig)

	namespaces, err := util.PopulateNamespaceList(kubeClient, appConfig.NamespaceSelector)

	if err != nil {
		logger.Error("An error occurred while populating namespaces", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Debug("Namespaces to look for hajimari apps: ", namespaces)
	hajimariApps, err = ingressAppsList.Populate(namespaces...).Get()

	if err != nil {
		logger.Error("An error occurred while looking for hajimari apps", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	customAppsList := customapps.NewList(*appConfig)

	customHajimariApps, err := customAppsList.Populate().Get()
	if err != nil {
		logger.Error("An error occured while populating custom hajimari apps", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// Append both generated and custom apps
	hajimariApps = append(hajimariApps, customHajimariApps...)

	w.Header().Add("Content-Type", "text/html")

	tplerr := p.tpl.Execute(w, struct {
		Title     string
		Greeting  string
		Date      string
		Apps      []hajimari.App
		Groups    []config.Group
		Providers []config.Provider
		Modules   []config.Module
	}{
		Title:     appConfig.Title,
		Greeting:  greet(appConfig.Name, time.Now().Hour()),
		Date:      time.Now().Format("Mon, Jan 02"),
		Apps:      hajimariApps,
		Groups:    appConfig.Groups,
		Providers: appConfig.Providers,
		Modules:   appConfig.Modules,
	})

	if tplerr != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// greet returns the greeting to be used in the h1 heading
func greet(name string, currentHour int) (greet string) {
	switch currentHour / 6 {
	case 0:
		greet = "Good night"
	case 1:
		greet = "Good morning"
	case 2:
		greet = "Good afternoon"
	default:
		greet = "Good evening"
	}

	if name != "" {
		return fmt.Sprintf("%s, %s!", greet, name)
	}

	return fmt.Sprintf("%s!", greet)
}