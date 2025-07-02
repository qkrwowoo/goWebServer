import requests

url = "http://127.0.0.1:8080/register"
headers = {
    "Content-Type": "application/json"
}
data = {
    "UserID": "testuser",
    "UserPW": "newpassword123",
    "DbType": "mysql"
}

response = requests.post(url, json=data, headers=headers)

print("Status Code:", response.status_code)
print("Response Body:", response.text)
