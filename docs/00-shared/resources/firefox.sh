# Crear el directorio para las claves
sudo install -d -m 0755 /etc/apt/keyrings

# Importar la clave de Mozilla
wget -q https://packages.mozilla.org/apt/repo-signing-key.gpg -O- | \
    sudo gpg --dearmor -o /etc/apt/keyrings/packages.mozilla.org.gpg

# Añadir el repositorio
echo "deb [signed-by=/etc/apt/keyrings/packages.mozilla.org.gpg] \
https://packages.mozilla.org/apt mozilla main" | \
sudo tee /etc/apt/sources.list.d/mozilla.list

# Dar prioridad al repositorio de Mozilla
echo '
Package: *
Pin: origin packages.mozilla.org
Pin-Priority: 1000
' | sudo tee /etc/apt/preferences.d/mozilla

# Actualizar
sudo apt update

# Instalar Firefox
sudo apt install firefox