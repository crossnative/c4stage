package c4

import (
	"io"
	"log"
	"text/template"
)

const PLANT_UML_TPL_C4_CONTEXT = `
@startuml
!include <C4/C4_Context>

SHOW_PERSON_PORTRAIT()

AddElementTag("experimental", $bgColor="DeepSkyBlue")
AddElementTag("deprecated", $bgColor="DarkCyan")

{{- range .Persons}}
Person({{.ID}}, "{{.Title}}", "{{.Description}}")
{{- end}}

{{- range .ExternalSystems}}
System_Ext({{.ID}}, "{{.Title}}")
{{- end}}

{{- range .Systems}}
System({{.ID}}, "{{.Title}}", "{{.Description}}", $tags="{{.AsTags}}")
{{- end}}

{{- range .Relations}}
Rel({{.SourceID}}, {{.TargetID}}, "{{.Label}}")
{{- end}}

SHOW_LEGEND()

@enduml
`

const PLANT_UML_TPL_C4_LANDSCAPE_CONTAINER = `
@startuml
!include <C4/C4_Container>
!include <tupadr3/common>
!include <tupadr3/devicons/java>
!include <tupadr3/devicons/angular>
!include <tupadr3/devicons2/spring>
!include <tupadr3/devicons2/c>
!include <tupadr3/devicons2/oracle_original>

SHOW_PERSON_PORTRAIT()

AddElementTag("experimental", $bgColor="DeepSkyBlue")
AddElementTag("deprecated", $bgColor="DarkCyan")

{{- range .ExternalSystems}}
System_Ext({{.ID}}, "{{.Title}}", "{{.Description}}")
{{- end}}

{{- range .Persons}}
Person({{.ID}}, "{{.Title}}", "{{.Description}}")
{{- end}}

{{- range .Systems}}
System_Boundary({{.ID}}, "{{.Title}}", "{{.Description}}") {
	{{- range .Containers}}
		{{- if .IsDatabase }}
		ContainerDb({{.ID}}, "{{.Title}}", "{{.Label}}", "{{.Description}}", $tags="{{.AsTags}}", $sprite="{{.Sprite}}")
		{{- else if .IsPerson }}
		{{- else }}
		Container({{.ID}}, "{{.Title}}", "{{.Label}}", "{{.Description}}", $tags="{{.AsTags}}", $sprite="{{.Sprite}}")
		{{- end }}
	{{- end}}
}
{{- end}}

{{- range .Relations}}
Rel({{.SourceID}}, {{.TargetID}}, "{{.Label}}")
{{- end}}

SHOW_LEGEND()

@enduml
`

type plantUMLExporter struct {
	templateContext   *template.Template
	templateContainer *template.Template
}

func newPlantUMLExporter() *plantUMLExporter {
	e := &plantUMLExporter{}

	t, err := template.New("").Parse(PLANT_UML_TPL_C4_CONTEXT)
	if err != nil {
		log.Fatal(err)
	}
	e.templateContext = t

	t, err = template.New("").Parse(PLANT_UML_TPL_C4_LANDSCAPE_CONTAINER)
	if err != nil {
		log.Fatal(err)
	}
	e.templateContainer = t

	return e
}

func (e *plantUMLExporter) ExportToPlantUMLContainer(c4Model *C4DiagramModel, w io.Writer) error {
	return e.templateContainer.Execute(w, c4Model)
}

func (e *plantUMLExporter) ExportToPlantUMLContext(c4Model *C4DiagramModel, w io.Writer) error {
	return e.templateContext.Execute(w, c4Model)
}
