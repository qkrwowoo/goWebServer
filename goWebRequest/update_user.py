import requests

with open("token.txt", "r") as f:
    token = f.read().strip()

url = "http://127.0.0.1:8080/user/update"
headers = {
    "Content-Type": "application/json",
    "Authorization": f"Bearer {token}"
}
data = {
    "UserID": "testuser2",
    "UserPW": "newpassword123",
    "NewUserPw": "newpassword123",
    "DbType": "mysql"
}

response = requests.post(url, json=data, headers=headers)

print("Status Code:", response.status_code)
print("Response Body:", response.text)

