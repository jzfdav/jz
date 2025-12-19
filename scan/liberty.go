package scan

import (
	"encoding/xml"
	"io"
	"jz/model"
	"os"
)

// ScanLiberty parses a WebSphere Liberty server.xml file and extracts configuration.
func ScanLiberty(path string) (model.LibertyServer, error) {
	f, err := os.Open(path)
	if err != nil {
		return model.LibertyServer{}, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return model.LibertyServer{}, err
	}

	var xs xmlServer
	if err := xml.Unmarshal(data, &xs); err != nil {
		return model.LibertyServer{}, err
	}

	server := model.LibertyServer{
		Name:            xs.Name,
		ServerXML:       path,
		EnabledFeatures: make([]string, 0),
		DeployedApps:    make([]model.LibertyApp, 0),
	}

	// Helper to add unique features
	seenFeatures := make(map[string]bool)
	for _, fm := range xs.FeatureManagers {
		for _, feat := range fm.Features {
			if feat != "" && !seenFeatures[feat] {
				server.EnabledFeatures = append(server.EnabledFeatures, feat)
				seenFeatures[feat] = true
			}
		}
	}

	// Process Applications
	for _, app := range xs.Applications {
		server.DeployedApps = append(server.DeployedApps, model.LibertyApp{
			ID:       app.ID,
			Location: app.Location,
			Type:     "application",
		})
	}

	// Process WebApplications
	for _, webApp := range xs.WebApplications {
		server.DeployedApps = append(server.DeployedApps, model.LibertyApp{
			ID:          webApp.ID,
			Location:    webApp.Location,
			ContextRoot: webApp.ContextRoot,
			Type:        "webApplication",
		})
	}

	return server, nil
}

// XML mapping structs

type xmlServer struct {
	Name            string              `xml:"name,attr"`
	FeatureManagers []xmlFeatureManager `xml:"featureManager"`
	Applications    []xmlApplication    `xml:"application"`
	WebApplications []xmlWebApplication `xml:"webApplication"`
}

type xmlFeatureManager struct {
	Features []string `xml:"feature"`
}

type xmlApplication struct {
	ID       string `xml:"id,attr"`
	Location string `xml:"location,attr"`
}

type xmlWebApplication struct {
	ID          string `xml:"id,attr"`
	Location    string `xml:"location,attr"`
	ContextRoot string `xml:"contextRoot,attr"`
}
