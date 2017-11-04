<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">
    <title>{{if .New }}新しい活動記録{{else}}活動記録の編集{{end}}</title>
    <link rel="stylesheet" href="/static/css/edit.css">
  </head>
  <body>
    <h1>{{if .New }}新しい活動記録{{else}}活動記録の編集{{end}}</h1>

    {{if .Error }}
      <p class="error">{{ .Error }}</p>
    {{end}}

    <form action="/create" method="POST">
      <p>日付: <input type="text" name="date" value="{{ .Date.Format "20060102" }}"></p>
      <p>タイトル: <input type="text" name="title" value="{{ .Title }}"></p>
      <p><textarea name="body" cols="80" rows="20" placeholder="マークダウンで記述">{{ .Body }}</textarea></p>
      <p>パスワード: <input type="password" name="password"></p>
      <p><button type="submit">投稿</button></p>
    </form>
  </body>
</html>
