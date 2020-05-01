package puml

const c4ContextTemplate = `
@startuml
!include https://raw.githubusercontent.com/koderizer/arc/master/viz/puml/C4-PlantUML/C4_Context.puml

title {{.Title}} 
{{range .Arc.Users}}
Person({{.Name | CleanID}}, "{{.Name}}", "{{.Desc | CleanUp}}")
{{end}}

Enterprise_Boundary({{.Arc.App}}, "{{.Arc.Desc}}") {
{{range .Arc.InternalSystems}}
	System({{.Name | CleanID}}, "{{.Name}}","{{.Desc | CleanUp}}")
{{end}}
}
{{range .Arc.ExternalSystems}}
System_Ext({{.Name | CleanID}}, "{{.Name}}", "{{.Desc | CleanUp}}")
{{end}}
{{range .Relations}}
{{if (ne .PointerTech "")}}
Rel({{.Subject | CleanID}},{{.Object | CleanID}},"{{.Pointer}}","{{.PointerTech}}")
{{else}}
Rel({{.Subject | CleanID}},{{.Object | CleanID}},"{{.Pointer}}")
{{end}}
{{end}}
@enduml`
