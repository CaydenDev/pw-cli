[![Build and Test](https://github.com/CaydenDev/pw-cli/actions/workflows/build.yml/badge.svg)](https://github.com/CaydenDev/pw-cli/actions/workflows/build.yml)

Quick Install

1. Build the executable:
   ```powershell
   .\build.ps1
   ```

2. Install system-wide:
   ```powershell
   .\install.ps1
   ```

Manual Usage

run the executable directly from the `dist` folder after building:

```powershell
.\dist\pwvault.exe
```

File Locations

- When installed system-wide: `C:\Program Files\PasswordVault\pwvault.exe`
- Vault file location: `vault.dat` in the same directory as the executable
