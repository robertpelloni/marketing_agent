with open("/var/www/tormentnexus.site/index.html", "r") as f:
    html = f.read()

# Remove the entire RobotSkull class
old_skull_class = "        // Robot Skull class — floating skull with jaw animation\n        class RobotSkull {"
end_marker = "        // Drone class — creepy robot spider"

start = html.find(old_skull_class)
end = html.find(end_marker)

if start != -1 and end != -1:
    html = html[:start] + html[end:]
    print("RobotSkull class removed")
else:
    print("Could not find RobotSkull class")

# Remove skulls from entities array
html = html.replace(
    "const skulls = [];\n        for (let i = 0; i < 6; i++) skulls.push(new RobotSkull());\n        ",
    "",
)

# Remove skulls from animate loop
html = html.replace(
    "// Update and draw skulls\n            skulls.forEach(s => { s.update(); s.draw(); });\n\n            ",
    "",
)

with open("/var/www/tormentnexus.site/index.html", "w") as f:
    f.write(html)
print("All skull references removed")
