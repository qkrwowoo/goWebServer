import requests
import json

url = "http://15.165.161.150:8080/login"
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

