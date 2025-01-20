package templates

const TokenPageTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Generate Terraform Token</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 40px auto; padding: 0 20px; }
        .button { 
            background-color: #0077cc; 
            color: white; 
            padding: 10px 20px; 
            border: none; 
            border-radius: 4px; 
            cursor: pointer; 
        }
        .token { 
            background-color: #f5f5f5; 
            padding: 20px; 
            margin: 20px 0; 
            border-radius: 4px; 
            word-break: break-all;
        }
    </style>
</head>
<body>
    <h1>Generate Terraform Token</h1>
    <p>Click the button below to generate a new token for Terraform:</p>
    <form method="POST" action="/app/settings/tokens/create">
        <input type="hidden" name="code" value="{{.Code}}">
        <button type="submit" class="button">Generate Token</button>
    </form>
    {{if .Token}}
    <div class="token">
        <h3>Your Token:</h3>
        <p>{{.Token}}</p>
        <p><small>Please copy this token and paste it into your Terraform CLI.</small></p>
    </div>
    {{end}}
</body>
</html>
`
