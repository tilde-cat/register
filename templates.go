package main

import "html/template"

var header = template.Must(template.New("header").Parse(`<html>
<head>
<title>{{.Title}}</title>
<link rel="stylesheet" href="http://tilde.cat/style.css">
</head>
<body>
`))

var footer = template.Must(header.New("footer").Parse(`</body>
</html>`))

var statusTemplate = template.Must(footer.New("status").Parse(`
{{ template "header" .Global }}
Status: {{ .Status }}
{{ template "footer" .Global }}`))

var formTemplate = template.Must(header.New("form").Parse(`{{ template "header" . }}
<h1>~🐱 Sign up form</h1>
<form action="/post" method="post">
<table>
<tr>
<th>Username:</th>
<th><input type="text" name="username"><br/></th>
</tr>
<tr>
<th>Email:</th>
<th><input type="email" name="email"><br/></th>
</tr>
<tr>
<th>Why would you want an account here?</th>
<th><textarea name="why" cols=50 rows=10></textarea><br/></th>
</tr>
<tr>
<th>SSH key:</th>
<th><textarea name="sshpublickey" cols=50 rows=10></textarea><br/></th>
</tr>
<tr><th colspan="2"><input type="submit" value="Submit"></th></tr>
</table>
</form>
{{ template "footer" . }}
`))
