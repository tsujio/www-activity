<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">
    <title>活動記録</title>
    <link rel="stylesheet" href="/static/css/index.css">
  </head>
  <body>
    <h1>活動記録</h1>

    {{ range . }}
      <article>
        <p class="pubdate"><time datetime="{{ .Date }}">{{ .Date }}</time></p>
        {{ .Content }}
      </article>
    {{ end }}
  </body>
</html>
