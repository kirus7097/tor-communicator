import tkinter as tk

root = tk.Tk()
root.title("PLACE HOLDER")
root.geometry("600x400")
root.resizable(False, False)

placeholder = tk.Label(
    root,
    text="PLACE HOLDER",
    font=("Arial", 32, "bold")
)
placeholder.pack(expand=True)

root.mainloop()