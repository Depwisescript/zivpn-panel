# ⚡ ZiVPN Panel

Panel de administración terminal para el servidor **udp-zivpn**. Gestiona usuarios con vigencia directamente desde la terminal de Linux.

## 🚀 Instalación Rápida

```bash
bash <(curl -sL https://raw.githubusercontent.com/Depwisescript/zivpn-panel/main/install.sh)
```

## 📋 Características

- **Instalar/Desinstalar** ZiVPN con un solo clic
- **Crear usuarios** con fecha de expiración (1-365 días)
- **Eliminar usuarios** individuales
- **Listar usuarios** con tabla de estado y vigencia
- **Purga automática** de usuarios vencidos al abrir el panel
- **Enrutamiento UDP** automático (rango 6000-19999)
- **Certificados SSL** generados automáticamente

## 🖥️ Uso

Después de instalar, ejecuta:

```bash
zivpn-panel
```

Se abrirá un menú interactivo:

```
╔══════════════════════════════════════════════╗
║        ⚡ ZIVPN PANEL v1.0.0               ║
╠══════════════════════════════════════════════╣
║  Estado: ● Activo (Puerto: 7300)            ║
╠══════════════════════════════════════════════╣
║  1. 📥 Instalar ZiVPN                      ║
║  2. 🗑️  Desinstalar ZiVPN                   ║
║  3. 👤 Crear Usuario                        ║
║  4. ❌ Eliminar Usuario                     ║
║  5. 📋 Listar Usuarios                      ║
║  6. 🧹 Purgar Expirados                    ║
║  7. ℹ️  Estado del Servicio                  ║
║  0. 🚪 Salir                               ║
╚══════════════════════════════════════════════╝
```

## 📂 Archivos del Sistema

| Archivo | Descripción |
|---|---|
| `/usr/local/bin/zivpn` | Binario del servidor udp-zivpn |
| `/usr/local/bin/zivpn-panel` | Panel de administración |
| `/etc/zivpn/config.json` | Configuración del daemon |
| `/etc/zivpn/users.json` | Base de datos de usuarios con vigencia |
| `/etc/zivpn/zivpn.crt` | Certificado SSL |
| `/etc/zivpn/zivpn.key` | Llave privada SSL |

## 🏗️ Compilación Manual

Requisitos: Go 1.21+

```bash
git clone https://github.com/Depwisescript/zivpn-panel.git
cd zivpn-panel
go build -o zivpn-panel ./cmd/zivpn-panel/
sudo mv zivpn-panel /usr/local/bin/
```

## 📝 Notas

- Requiere **root** para ejecutarse
- Compatible con **amd64** y **arm64**
- Los usuarios expirados se eliminan automáticamente al abrir el panel
- El rango de puertos UDP **6000-19999** se redirige al puerto de ZiVPN
