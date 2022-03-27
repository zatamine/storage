//use item_storage

db.createUser(
  {
    user: "admin",
    pwd:  "pass",
    roles: [
			{ role: "readWrite", db: "item_storage" }
		]
  }
)
