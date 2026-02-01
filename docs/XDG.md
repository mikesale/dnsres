# XDG Base Directory Support

## Overview

dnsres follows the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html) for organizing configuration files, logs, and data. This provides a standardized, clean approach to file organization that integrates well with modern Unix-like systems.

## Directory Structure

### Configuration Directory
**Location:** `$XDG_CONFIG_HOME/dnsres/` (default: `~/.config/dnsres/`)

Contains:
- `config.json` - Main configuration file

**Creation:** Automatically created on first run with sensible defaults if no config file exists.

**Permissions:** 
- Directory: `0755` (drwxr-xr-x)
- Files: `0644` (-rw-r--r--)

### State Directory (Logs)
**Location:** `$XDG_STATE_HOME/dnsres/` (default: `~/.local/state/dnsres/`)

Contains:
- `dnsres-success.log` - Successful DNS resolutions
- `dnsres-error.log` - Failed DNS resolutions
- `dnsres-app.log` - Application lifecycle events

**Creation:** Automatically created when dnsres starts.

**Fallback:** If the XDG state directory cannot be created (permissions, disk full, etc.), dnsres will fall back to `$HOME/logs/` and notify the user.

**Permissions:**
- Directory: `0755` (drwxr-xr-x)
- Files: `0644` (-rw-r--r--)

### Data Directory
**Location:** `$XDG_DATA_HOME/dnsres/` (default: `~/.local/share/dnsres/`)

**Status:** Reserved for future use (e.g., persistent DNS cache, historical data).

## Configuration File Discovery

dnsres searches for configuration files in the following order:

1. **Explicit path** (highest priority)
   - Via `-config` flag: `dnsres -config /path/to/config.json`
   - Always used if provided, no further searching

2. **Current directory** (backward compatibility)
   - `./config.json`
   - Checked if no explicit path provided
   - Maintains compatibility with existing deployments

3. **XDG config directory** (modern standard)
   - `$XDG_CONFIG_HOME/dnsres/config.json`
   - Typically `~/.config/dnsres/config.json`
   - Created automatically if missing

4. **Built-in defaults** (fallback)
   - If no config file found anywhere
   - Uses hardcoded defaults for all settings

## Automatic File Creation

### Config File Auto-Creation

When dnsres cannot find a configuration file, it automatically creates one at `~/.config/dnsres/config.json` with the following defaults:

```json
{
  "cache": {
    "max_size": 1000
  },
  "circuit_breaker": {
    "threshold": 5,
    "timeout": "30s"
  },
  "dns_servers": [
    "8.8.8.8:53",
    "1.1.1.1:53",
    "9.9.9.9:53"
  ],
  "health_port": 8880,
  "hostnames": [
    "example.com"
  ],
  "instrumentation_level": "none",
  "log_dir": "",
  "metrics_port": 9990,
  "query_interval": "30s",
  "query_timeout": "5s"
}
```

**What gets created:**
- Complete, valid configuration file
- All required and optional fields
- Ready to use immediately
- Can be customized by editing the file

**User notification:**
```
Loading configuration from /Users/username/.config/dnsres/config.json
Created default configuration file at /Users/username/.config/dnsres/config.json
```

### Log Directory Auto-Creation

dnsres automatically creates the log directory when starting:

1. **Attempts XDG state directory first:**
   - Creates `~/.local/state/dnsres/`
   - Creates parent directories if needed (`~/.local/state/`)
   - No user notification if successful

2. **Falls back to `$HOME/logs` if XDG fails:**
   - Creates `$HOME/logs/`
   - User is notified:
     ```
     Note: Using fallback log directory at /Users/username/logs
     (XDG state directory unavailable)
     ```

3. **Fails with error if both fail:**
   - Reports inability to create log directory
   - Application exits

## Environment Variables

dnsres respects standard XDG environment variables for customizing file locations.

### XDG_CONFIG_HOME

Override the default config directory:

```bash
export XDG_CONFIG_HOME="$HOME/.my-config"
dnsres example.com
# Config will be at: $HOME/.my-config/dnsres/config.json
```

**Default:** `~/.config`

### XDG_STATE_HOME

Override the default log directory:

```bash
export XDG_STATE_HOME="$HOME/.my-state"
dnsres example.com
# Logs will be at: $HOME/.my-state/dnsres/*.log
```

**Default:** `~/.local/state`

### XDG_DATA_HOME

Override the default data directory (currently unused):

```bash
export XDG_DATA_HOME="$HOME/.my-data"
```

**Default:** `~/.local/share`

### Combined Example

```bash
export XDG_CONFIG_HOME="$HOME/config"
export XDG_STATE_HOME="$HOME/state"
export XDG_DATA_HOME="$HOME/data"

dnsres example.com

# Files will be at:
# - Config: $HOME/config/dnsres/config.json
# - Logs: $HOME/state/dnsres/*.log
# - Data: $HOME/data/dnsres/ (future use)
```

## Fallback Behavior

### Config File Fallback

If XDG config file creation fails:
- No error is raised
- Application uses built-in defaults
- User can still override with CLI arguments

### Log Directory Fallback

If XDG state directory creation fails:
- Falls back to `$HOME/logs/`
- User is notified via stdout (CLI) or TUI status message
- Application logs the fallback in app log

If `$HOME/logs/` also fails:
- Application exits with error
- Error message indicates inability to create log directory

## Backward Compatibility

dnsres maintains full backward compatibility with pre-XDG versions:

### Existing Config Files

If you have `./config.json` in your current directory:
- It will be used automatically
- No XDG config will be created or used
- No changes required to existing workflows

### Existing Log Directories

If you have `log_dir` specified in your config:
- Your specified directory is always used
- XDG directories are ignored
- No migration or changes required

### Explicit Paths

All explicit paths continue to work:
```bash
dnsres -config /etc/dnsres/config.json
# Uses exactly this path, no XDG involvement
```

## Migration Guide

### For New Users

**No action required!** Just run dnsres:

```bash
dnsres example.com
```

Everything is created automatically.

### For Existing Users

You have three options:

#### Option 1: Keep Current Setup (Recommended)
- No changes needed
- Continue using `./config.json` and `./logs/`
- XDG support is transparent

#### Option 2: Migrate to XDG
1. Move your config:
   ```bash
   mkdir -p ~/.config/dnsres
   mv config.json ~/.config/dnsres/
   ```

2. Update config to use XDG logs:
   ```bash
   # Edit ~/.config/dnsres/config.json
   # Set "log_dir": "" or remove it entirely
   ```

3. Optional: Move existing logs:
   ```bash
   mkdir -p ~/.local/state/dnsres
   mv logs/*.log ~/.local/state/dnsres/
   ```

#### Option 3: Hybrid Approach
- Keep config in `./config.json` (backward compatibility)
- Use XDG for logs (set `"log_dir": ""` in config)

## Benefits of XDG Support

### For Users

✓ **Clean home directory** - No dot-files or directories cluttering `$HOME`  
✓ **Organized system** - All config in `~/.config/`, all logs in `~/.local/state/`  
✓ **Discoverable** - Standard locations make files easy to find  
✓ **Backup-friendly** - Backup tools know where to look  
✓ **Multi-user friendly** - Each user has separate config/logs  

### For System Administrators

✓ **Standard compliance** - Follows modern Unix conventions  
✓ **Scriptable** - Can control via environment variables  
✓ **Auditable** - Predictable file locations  
✓ **Cleanup-friendly** - Easy to identify and clean old logs  

### For Developers

✓ **Testable** - Can redirect to temp directories via env vars  
✓ **Containerizable** - Easy to mount specific directories  
✓ **Debuggable** - Logs have consistent, predictable locations  

## Technical Implementation

### Path Resolution Algorithm

**Config file:**
```
1. if -config flag provided:
     return flag value
2. if ./config.json exists:
     return "./config.json"
3. attempt $XDG_CONFIG_HOME/dnsres/config.json:
     if exists: return it
     else: create with defaults and return it
4. if creation fails:
     return "" (use built-in defaults)
```

**Log directory:**
```
1. if log_dir in config is non-empty:
     return config value
2. attempt $XDG_STATE_HOME/dnsres:
     if created successfully: return it
     else: fall back to $HOME/logs
3. if fallback created successfully:
     notify user and return it
4. if fallback fails:
     exit with error
```

### Code Organization

XDG support is implemented in `internal/xdg/xdg.go`:

- `ConfigHome()` - Returns XDG config directory
- `StateHome()` - Returns XDG state directory  
- `DataHome()` - Returns XDG data directory
- `ConfigFile()` - Returns config file path, creates if needed
- `EnsureStateDir()` - Returns log directory path with fallback

## Troubleshooting

### Config file not being created

**Symptoms:**
- No config file at `~/.config/dnsres/config.json`
- dnsres uses built-in defaults

**Possible causes:**
- `./config.json` exists in current directory (will be used instead)
- `-config` flag is being used
- Permissions issue preventing directory creation

**Solutions:**
```bash
# Check if local config exists
ls -la config.json

# Check config directory permissions
ls -ld ~/.config

# Try creating manually
mkdir -p ~/.config/dnsres
dnsres example.com
```

### Logs going to unexpected location

**Symptoms:**
- Logs not in `~/.local/state/dnsres/`
- Fallback message shown

**Possible causes:**
- XDG state directory cannot be created
- `log_dir` is set in config
- Permissions issue

**Solutions:**
```bash
# Check state directory permissions
ls -ld ~/.local/state

# Check config for log_dir setting
cat ~/.config/dnsres/config.json | grep log_dir

# Create state directory manually
mkdir -p ~/.local/state/dnsres
chmod 755 ~/.local/state/dnsres
```

### Permission denied errors

**Symptoms:**
- Cannot create config or log directories
- Application exits with permission errors

**Solutions:**
```bash
# Fix config directory permissions
chmod 755 ~/.config
mkdir -p ~/.config/dnsres
chmod 755 ~/.config/dnsres

# Fix state directory permissions
mkdir -p ~/.local/state
chmod 755 ~/.local/state
mkdir -p ~/.local/state/dnsres
chmod 755 ~/.local/state/dnsres
```

## Testing XDG Support

### Test Config Auto-Creation

```bash
# Clean slate
rm -rf ~/.config/dnsres

# Run dnsres
dnsres example.com

# Verify
ls -la ~/.config/dnsres/config.json
cat ~/.config/dnsres/config.json
```

### Test Log Directory Creation

```bash
# Clean slate
rm -rf ~/.local/state/dnsres

# Run dnsres
dnsres example.com

# Verify
ls -la ~/.local/state/dnsres/
```

### Test Fallback Behavior

```bash
# Make state directory unavailable
chmod 000 ~/.local/state

# Run dnsres (will use fallback)
dnsres example.com

# Check for fallback message and logs
ls -la ~/logs/

# Restore permissions
chmod 755 ~/.local/state
```

### Test Environment Variables

```bash
# Use custom directories
export XDG_CONFIG_HOME="/tmp/my-config"
export XDG_STATE_HOME="/tmp/my-state"

dnsres example.com

# Verify
ls -la /tmp/my-config/dnsres/
ls -la /tmp/my-state/dnsres/
```

## Related Documentation

- [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html)
- [README.md](../README.md) - Main documentation
- [DEVELOPMENT.md](DEVELOPMENT.md) - Development guide
- [ARCHITECTURE.md](ARCHITECTURE.md) - Architecture overview
