import tkinter as tk
from tkinter import ttk, simpledialog, messagebox
import requests

API_URL = "http://127.0.0.1:8080/addcontact"

response = requests.get(
    API_URL,
    timeout=5
)

if response.status_code != 200:
    raise RuntimeError("Could not check login status")

body = response.json()
data = body.get("data", {})

if not data.get("loggedIn"):
    raise RuntimeError("User is not logged in")

username = data.get("username")

if not username:
    raise RuntimeError("Username is missing")

current_contact = None  # z kim aktualnie rozmawiamy

contacts = []


def select_contact(name: str):
    #select current contact and update the chat window
    global current_contact
    current_contact = name
    message_with_var.set(f"Message with {name}")
    chat_text.configure(state="normal")
    chat_text.delete("1.0", tk.END)
    chat_text.insert(tk.END, f"--- Talk with {name} ---\n")
    chat_text.configure(state="disabled")


def send_message(event=None):
    #sends messages to currently picked chat
    if current_contact is None:
        return

    text = message_entry.get().strip()
    if not text:
        return

    chat_text.configure(state="normal")
    chat_text.insert(tk.END, f"{username}: {text}\n")
    chat_text.configure(state="disabled")
    chat_text.see(tk.END)

    message_entry.delete(0, tk.END)

root = tk.Tk()
root.title("Torchat - Chat")
root.geometry("1500x1000")
root.resizable(False, False)

root.columnconfigure(0, weight=1)
root.rowconfigure(1, weight=1)

# ---"Message with ..." -------------------------------------
top_bar = tk.Frame(root, height=60, bg="#2c2f33")
top_bar.grid(row=0, column=0, sticky="nsew")
top_bar.grid_propagate(False)

message_with_var = tk.StringVar(value="Message with ...")
message_with_label = tk.Label(
    top_bar,
    textvariable=message_with_var,
    font=("Arial", 18, "bold"),
    bg="#2c2f33",
    fg="white",
)
message_with_label.pack(side="left", padx=20, pady=10)


main_area = tk.Frame(root)
main_area.grid(row=1, column=0, sticky="nsew")
main_area.columnconfigure(0, weight=1)
main_area.columnconfigure(1, weight=0)
main_area.rowconfigure(0, weight=1)

#chat panel
chat_frame = tk.Frame(main_area)
chat_frame.grid(row=0, column=0, sticky="nsew", padx=(10, 5), pady=10)
chat_frame.rowconfigure(0, weight=1)
chat_frame.columnconfigure(0, weight=1)

chat_scroll = tk.Scrollbar(chat_frame)
chat_scroll.grid(row=0, column=1, sticky="ns")

chat_text = tk.Text(
    chat_frame,
    state="disabled",
    wrap="word",
    yscrollcommand=chat_scroll.set,
    font=("Arial", 12),
)
chat_text.grid(row=0, column=0, sticky="nsew")
chat_scroll.config(command=chat_text.yview)

# write and send messages 
input_frame = tk.Frame(chat_frame)
input_frame.grid(row=1, column=0, columnspan=2, sticky="ew", pady=(10, 0))
input_frame.columnconfigure(0, weight=1)

message_entry = tk.Entry(input_frame, font=("Arial", 12))
message_entry.grid(row=0, column=0, sticky="ew", padx=(0, 5))
message_entry.bind("<Return>", send_message)

send_button = tk.Button(input_frame, text="Send", command=send_message)
send_button.grid(row=0, column=1)

# --- username and contacts ---------
sidebar = tk.Frame(main_area, width=280, bg="#23272a")
sidebar.grid(row=0, column=1, sticky="nsew", padx=(5, 10), pady=10)
sidebar.grid_propagate(False)

username_label = tk.Label(
    sidebar,
    text=username,
    font=("Arial", 16, "bold"),
    bg="#23272a",
    fg="white",
    anchor="e",
    justify="right",
)
username_label.pack(fill="x", padx=10, pady=(10, 5))

separator = ttk.Separator(sidebar, orient="horizontal")
separator.pack(fill="x", padx=10, pady=5)

contacts_header = tk.Frame(sidebar, bg="#23272a")
contacts_header.pack(fill="x", padx=10, pady=(5, 0))
contacts_header.columnconfigure(0, weight=1)

contacts_label = tk.Label(
    contacts_header,
    text="Contacts",
    font=("Arial", 12, "bold"),
    bg="#23272a",
    fg="#b9bbbe",
    anchor="w",
)
contacts_label.grid(row=0, column=0, sticky="w")

add_contact_button = tk.Button(
    contacts_header,
    text="+",
    font=("Arial", 12, "bold"),
    bg="#7289da",
    fg="white",
    activebackground="#5b6eae",
    activeforeground="white",
    bd=0,
    width=3,
    cursor="hand2",
    command=lambda: add_contact(),
)
add_contact_button.grid(row=0, column=1, sticky="e")

add_remove_messages_button = tk.Button(
    contacts_header,
    text="Remove messages",
    font=("Arial", 10),
    bg="#f21b0c",
    fg = "white",
    activebackground="#d10f00",
    activeforeground="white",
    bd=0,
    width=15,
    cursor="hand2",
    command=lambda: remove_messages(),
)
add_remove_messages_button.grid(row=1, column=0, columnspan=2, sticky="ew", pady=(5, 0))
contacts_list = tk.Listbox(
    sidebar,
    font=("Arial", 12),
    bg="#2c2f33",
    fg="white",
    selectbackground="#7289da",
    activestyle="none",
    highlightthickness=0,
    bd=0,
)
contacts_list.pack(fill="both", expand=True, padx=10, pady=10)

for contact in contacts:
    contacts_list.insert(tk.END, contact)


def on_contact_select(event):
    selection = contacts_list.curselection()
    if selection:
        name = contacts_list.get(selection[0])
        select_contact(name)


def remove_messages():
    delete_mssages = messagebox.askquestion(
        "Remove messages",
        "Are you sure you want to remove all messages?",
        icon="warning",
        parent=root,
    )

    if delete_mssages == "yes":
        messagebox.showinfo("Info", "All messages have been removed.", parent=root)
    if delete_mssages == "no":
        pass

def add_contact():
    name = simpledialog.askstring(
        "New Contact",
        "Enter the username:",
        parent=root,
    )
    if not name:
        return

    name = name.strip()
    if not name:
        return
    try:
        response = requests.post(
            API_URL,
            json={"contact": name},
            timeout=5,
        )
        if response.status_code != 200:
            print(response.status_code)
            print(response.text)
            raise RuntimeError("Could not add contact")
    except Exception as e:
        messagebox.showerror("Error", str(e), parent=root)
        return

    if name in contacts:
        messagebox.showinfo("Info", "This contact already exists.", parent=root)
        return

    contacts.append(name)
    contacts_list.insert(tk.END, name)

    select_contact(name)


contacts_list.bind("<<ListboxSelect>>", on_contact_select)

root.mainloop()