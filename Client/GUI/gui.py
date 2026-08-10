import sys
import os
import tkinter as tk
from tkinter import messagebox
import requests
import subprocess

API_URL = "http://127.0.0.1:8080/api"


def send_request(data):
    try:
        response = requests.post(
            API_URL,
            json=data,
            timeout=5
        )

        print("HTTP:", response.status_code)
        print("BODY:", response.text)

        return response.json()

    except requests.exceptions.ConnectionError:
        messagebox.showerror(
            "Connection error",
            "Server is not running"
        )
        return None

    except requests.exceptions.Timeout:
        messagebox.showerror(
            "Timeout",
            "Server did not respond"
        )
        return None

    except Exception as e:
        messagebox.showerror(
            "Error",
            str(e)
        )
        return None


def login():
    username = name_entry.get()
    password = password_entry.get()

    if username == "" or password == "":
        messagebox.showwarning(
            "Missing data",
            "Enter username and password"
        )
        return

    data = {
        "type": "login",
        "username": username,
        "password": password
    }

    response = send_request(data)

    if response is None:
        return

    if response.get("status") == "error":
        messagebox.showerror(
            "Login failed",
            response.get("message")
        )
    else:
        messagebox.showinfo(
            "Login",
            str(response.get("data"))
        )
        script_path = os.path.join(os.path.dirname(__file__), "chat.py")
        subprocess.Popen([sys.executable, script_path])
        window.destroy()


def register():
    username = name_entry.get()
    password = password_entry.get()

    if username == "" or password == "":
        messagebox.showwarning(
            "Missing data",
            "Enter username and password"
        )
        return

    data = {
        "type": "register",
        "username": username,
        "password": password
    }

    response = send_request(data)

    if response is None:
        return

    if response.get("status") == "error":
        messagebox.showerror(
            "Register failed",
            response.get("message")
        )
    else:
        messagebox.showinfo(
            "Register",
            str(response.get("data"))
        )
        script_path = os.path.join(os.path.dirname(__file__), "chat.py")
        subprocess.Popen([sys.executable, script_path])
        window.destroy()
window = tk.Tk()

window.title("TorChat")
window.geometry("700x400")
window.resizable(False, False)

title = tk.Label(
    window,
    text="TorChat Login / Register",
    font=("Arial", 14)
)

title.pack(pady=15)

tk.Label(
    window,
    text="Username"
).pack()

name_entry = tk.Entry(
    window,
    width=30
)

name_entry.pack()

tk.Label(
    window,
    text="Password"
).pack(pady=5)

password_entry = tk.Entry(
    window,
    width=30,
    show="*"
)

password_entry.pack()

button_frame = tk.Frame(window)

button_frame.pack(pady=20)

login_button = tk.Button(
    button_frame,
    text="Login",
    width=12,
    command=login
)

login_button.grid(
    row=0,
    column=0,
    padx=10
)

register_button = tk.Button(
    button_frame,
    text="Register",
    width=12,
    command=register
)

register_button.grid(
    row=0,
    column=1,
    padx=10
)

window.mainloop()