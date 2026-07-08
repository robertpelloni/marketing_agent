with open("internal/sales/order_processor_test.go", "r") as f:
    content = f.read()

content = content.replace("billing.Tier", "string")

with open("internal/sales/order_processor_test.go", "w") as f:
    f.write(content)
