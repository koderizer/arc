@startuml
!include https://raw.githubusercontent.com/koderizer/arc/master/viz/puml/C4-PlantUML/C4_Context.puml

title {{.Title}} 
{{range .Arc.Users}}
Person({{.Name | CleanID}}, "{{.Name}}", "{{.Desc | CleanUp}}")
{{end}}

{{range .Arc.InternalSystems}}
System({{.Name | CleanID}}, "{{.Name}}","{{.Desc | CleanUp}}")
{{end}}

{{range .Arc.ExternalSystems}}
System_Ext({{.Name | CleanID}}, "{{.Name}}", "{{.Desc | CleanUp}}")
{{end}}
{{range .Arc.Relations}}
# Rel({{.Subject}},{{.Object}},"{{.Path}}","{{.PathProperty}}")
{{end}}
@enduml