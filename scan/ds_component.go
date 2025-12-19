package scan

import (
	"encoding/xml"
	"io"
	"jz/model"
	"os"
)

// ScanDSComponents parses a list of Service-Component XML files and returns their metadata.
func ScanDSComponents(paths []string) ([]model.DSComponent, error) {
	var results []model.DSComponent

	for _, path := range paths {
		comp, err := parseDSFile(path)
		// We ignore file read/parse errors to correctly handle cases where files might be missing or invalid
		// but we still want to proceed with the others.
		// Ignore individual file errors and continue processing other components.
		if err == nil {
			results = append(results, comp)
		}
	}

	return results, nil
}

// xmlComponent is a helper struct for XML unmarshalling
type xmlComponent struct {
	Name           string            `xml:"name,attr"`
	Immediate      bool              `xml:"immediate,attr"`
	Implementation xmlImplementation `xml:"implementation"`
	Service        xmlService        `xml:"service"`
	References     []xmlReference    `xml:"reference"`
}

type xmlImplementation struct {
	Class string `xml:"class,attr"`
}

type xmlService struct {
	Provide []xmlProvide `xml:"provide"`
}

type xmlProvide struct {
	Interface string `xml:"interface,attr"`
}

type xmlReference struct {
	Interface string `xml:"interface,attr"`
}

func parseDSFile(path string) (model.DSComponent, error) {
	f, err := os.Open(path)
	if err != nil {
		return model.DSComponent{}, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return model.DSComponent{}, err
	}

	var xc xmlComponent
	if err := xml.Unmarshal(data, &xc); err != nil {
		return model.DSComponent{}, err
	}

	// Map to model.DSComponent
	comp := model.DSComponent{
		Name:                 xc.Name,
		ImplementationClass:  xc.Implementation.Class,
		Immediate:            xc.Immediate,
		SourceXML:            path,
		ProvidedInterfaces:   make([]string, 0),
		ReferencedInterfaces: make([]string, 0),
	}

	// Extract Provided Interfaces
	for _, p := range xc.Service.Provide {
		if p.Interface != "" {
			comp.ProvidedInterfaces = append(comp.ProvidedInterfaces, p.Interface)
		}
	}

	// Extract Referenced Interfaces
	for _, r := range xc.References {
		if r.Interface != "" {
			comp.ReferencedInterfaces = append(comp.ReferencedInterfaces, r.Interface)
		}
	}

	return comp, nil
}
