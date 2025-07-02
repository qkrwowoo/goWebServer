import requests

url = "http://13.125.250.202:8080/register"
headers = {
    "Content-Type": "application/json"
}
data = {
    "UserID": "testuser",
    "UserPW": "password123",
    "DbType": "mysql"
}

response = requests.post(url, json=data, headers=headers)

print("Status Code:", response.status_code)
print("Response Body:", response.text)
