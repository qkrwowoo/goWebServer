import requests

url = "http://15.165.161.150:8080/register"
headers = {
    "Content-Type": "application/json"
}
data = {
    "UserID": "testuser1",
    "UserPW": "password123",
    "DbType": "mysql"
}

response = requests.post(url, json=data, headers=headers)

print("Status Code:", response.status_code)
print("Response Body:", response.text)
