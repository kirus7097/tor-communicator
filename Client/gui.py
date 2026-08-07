import tkinter as tk
import requests
import json

def get_response():
    r = requests.get("http://127.0.0.1:8080/api")
    print(r.status_code)
    print(r.text)
    return r.json()

def login():
    global loginName, password
    loginName = name_entry.get()
    password = password_entry.get()
    print("Login name:", loginName)
    print("Password:", password)

    send_login = {
        "type": "login",
        "username": loginName,
        "password": password
    }

    try:
        requests.post("http://127.0.0.1:8080/api", json=send_login)
    except requests.exceptions.ConnectionError:
        print("Server is not running")

    get_response()

def register():
    global registerName, password
    registerName = name_entry.get()
    password = password_entry.get()
    print("Register name:", registerName)
    print("Password:", password)

    send_register = {
        "type": "register",
        "username": registerName,
        "password": password
    }

    try:
        requests.post("http://127.0.0.1:8080/api", json=send_register)
    except requests.exceptions.ConnectionError:
        print("Server is not running")
    get_response()

window = tk.Tk()
window.title("Login / Register")
window.resizable(False, False)
window.geometry("500x300")

tk.Label(window, text="Login / Register").pack(pady=5)

name_entry = tk.Entry(window, width=30)
name_entry.pack()

tk.Label(window, text="Password").pack(pady=5)

password_entry = tk.Entry(window, width=30, show="*")
password_entry.pack()

button_frame = tk.Frame(window)
button_frame.pack(pady=20)

login_button = tk.Button(button_frame, text="Login", width=10, command=login)
login_button.grid(row=0, column=0, padx=10)

register_button = tk.Button(button_frame, text="Register", width=10, command=register)
register_button.grid(row=0, column=1, padx=10)

window.mainloop()
