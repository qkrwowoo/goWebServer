import requests

url = "http://13.125.250.202:8080/user/update"
headers = {
    "Content-Type": "application/json",
    "Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTE0NTU4NTUsInVzZXJuYW1lIjoidGVzdHVzZXIifQ.wI6efYl9txKqCnN3W4UhebhC9zgrHGA8jxCj3ztn1W8"
}
data = {
    "UserID": "testuser",
    "UserPW": "password123",
    "NewUserPw": "newpassword123",
    "DbType": "mysql"
}

response = requests.post(url, json=data, headers=headers)

print("Status Code:", response.status_code)
print("Response Body:", response.text)

