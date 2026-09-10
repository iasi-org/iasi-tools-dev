# IASI Shiny

Extensión Quarto para incrustar aplicaciones Shiny para R que se ejecutan en
el navegador mediante Shinylive y webR. La salida HTML es estática y puede
publicarse en GitHub Pages.

## Requisitos de renderizado

```r
install.packages("shinylive")
```

El paquete solo es necesario al renderizar. El visitante de la página no
necesita R ni un servidor Shiny.

## Configuración

```yaml
filters:
  - iasi-shiny
```

## Uso

````qmd
```{shinylive-r}
#| standalone: true
#| viewerHeight: 450

library(shiny)

ui <- fluidPage(
  sliderInput("n", "Valor", 1, 100, 50),
  textOutput("value")
)

server <- function(input, output, session) {
  output$value <- renderText(input$n)
}

shinyApp(ui, server)
```
````

No debe utilizarse `embed-resources: true`, porque Shinylive necesita publicar
sus recursos web por separado.

La implementación adapta la extensión oficial `quarto-ext/shinylive`,
distribuida bajo licencia MIT.
