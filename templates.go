package main

import "html/template"

var formTemplate = template.Must(template.New("form").Parse(`
<html>
<body>
<h1>~🐱 signup form</h1>
<form action="/post" method="post">
Username:
<input type="text" name="username"><br/>
Email:
<input type="email" name="email"><br/>
Why would you want an account here?
<textarea name="why" cols=50 rows=10></textarea><br/>
SSH key:
<textarea name="sshpublickey" cols=50 rows=10></textarea><br/>
<input type="submit" value="Submit">
</form>
</body>
</html>
`))
