<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">
    <title>活動記録</title>
    <link rel="stylesheet" href="/static/css/index.css">
  </head>
  <body>
    <h1>活動記録</h1>

    <div>
      <a href="/new" class="new-activity-link">新しい活動記録</a>
      <div class="clearfix"></div>
    </div>

    {{ range . }}
      <article>
        <p class="pubdate"><time datetime="{{ .Date }}">{{ .Date }}</time></p>
        {{ .Content }}
      </article>
    {{ end }}
  </body>
</html>
