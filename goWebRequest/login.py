import requests

url = "http://127.0.0.1:8080/login"
headers = {
    "Content-Type": "application/json"
}
data = {
    "UserID": "testuser",
    "UserPW": "newpassword123",
    "DbType": "mysql"
}

response = requests.post(url, headers=headers, json=data)


if response.status_code == 200:
    res_json = response.json()
    token = res_json.get("message")
    if token:
        with open("token.txt", "w") as f:
            f.write(token)
    else:
        print("not found message")
print("Status Code:", response.status_code)
print("Response Body:", response.text)

