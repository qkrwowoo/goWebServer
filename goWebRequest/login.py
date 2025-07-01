import requests
import json

url = "http://localhost:8080/login"
headers = {
    "Content-Type": "application/json"
}
data = {
    "UserID": "testuser2",
    "UserPW": "newpassword123",
    "DbType": "mysql"
}

response = requests.post(url, headers=headers, data=json.dumps(data))

# 결과 출력
print("Status Code:", response.status_code)
print("Response Body:", response.text)

