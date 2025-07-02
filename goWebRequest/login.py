import requests

url = "http://15.165.161.150:8080/login"
#url = "http://127.0.0.1:8080/login"
headers = {
    "Content-Type": "application/json"
}
data = {
    "UserID": "testuser1",
    "UserPW": "password123",
    "DbType": "mysql"
}

response = requests.post(url, headers=headers, json=data)

print("Status Code:", response.status_code)
print("Response Body:", response.text)

