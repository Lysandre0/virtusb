#!/bin/bash

# virtusb - Virtual USB Device Manager
# Version: 1.0.0

set -euo pipefail

# Configuration

readonly SCRIPT_NAME="virtusb"
readonly SCRIPT_VERSION="1.0.0"

readonly GADGET_ROOT="/sys/kernel/config/usb_gadget"

# Set paths
readonly STATE_DIR="/opt/virtusb/data"
readonly IMAGE_DIR="/opt/virtusb/data/images"
readonly METADATA_DIR="/opt/virtusb/data/metadata"
readonly ENABLED_STATE_FILE="/opt/virtusb/data/enabled_devices.txt"

# Colors
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly NC='\033[0m'

# Logging

log_error() { echo -e "${RED}❌${NC} $1"; }
log_success() { echo -e "${GREEN}✅${NC} $1"; }
log_info() { echo -e "${YELLOW}ℹ️${NC} $1"; }

# Utilities

# State management
save_enabled_state() {
    local name="$1"
    local enabled_devices=()
    
    # Read existing enabled devices
    if [[ -f "$ENABLED_STATE_FILE" ]]; then
        while IFS= read -r device; do
            [[ -n "$device" ]] && enabled_devices+=("$device")
        done < "$ENABLED_STATE_FILE"
    fi
    
    # Add new device if not already present
    if [[ ! " ${enabled_devices[*]} " =~ " ${name} " ]]; then
        enabled_devices+=("$name")
    fi
    
    # Save updated list
    printf "%s\n" "${enabled_devices[@]}" > "$ENABLED_STATE_FILE"
}

remove_enabled_state() {
    local name="$1"
    local enabled_devices=()
    local updated_devices=()
    
    # Read existing enabled devices
    if [[ -f "$ENABLED_STATE_FILE" ]]; then
        while IFS= read -r device; do
            [[ -n "$device" ]] && enabled_devices+=("$device")
        done < "$ENABLED_STATE_FILE"
    fi
    
    # Remove the specified device
    for device in "${enabled_devices[@]}"; do
        [[ "$device" != "$name" ]] && updated_devices+=("$device")
    done
    
    # Save updated list
    printf "%s\n" "${updated_devices[@]}" > "$ENABLED_STATE_FILE"
}

get_enabled_devices() {
    local enabled_devices=()
    
    if [[ -f "$ENABLED_STATE_FILE" ]]; then
        while IFS= read -r device; do
            [[ -n "$device" ]] && enabled_devices+=("$device")
        done < "$ENABLED_STATE_FILE"
    fi
    
    echo "${enabled_devices[@]}"
}

restore_enabled_devices() {
    local enabled_devices
    read -ra enabled_devices <<< "$(get_enabled_devices)"
    
    if [[ ${#enabled_devices[@]} -eq 0 ]]; then
        return 0
    fi
    
    local restored_count=0
    local failed_count=0
    
    for device in "${enabled_devices[@]}"; do
        if check_gadget_integrity "$device"; then
            # Gadget exists and is complete, try to enable it
            enable_gadget "$device" "true" 2>/dev/null || {
                log_error "Échec de l'activation du gadget: $device"
                remove_enabled_state "$device"
                ((failed_count++))
            }
        else
            # Gadget missing but has metadata - try to recreate it
            if [[ -f "$METADATA_DIR/$device.meta" ]] && [[ -f "$IMAGE_DIR/$device.img" ]]; then
                log_info "Recréation du gadget après reboot: $device"
                
                # Read metadata to recreate gadget
                local vid_pid="" vendor="" product="" serial="" brand="" size=""
                while IFS='=' read -r key value; do
                    value="${value//\"/}"
                    case "$key" in
                        VID_PID) vid_pid="$value" ;;
                        VENDOR) vendor="$value" ;;
                        PRODUCT) product="$value" ;;
                        SERIAL) serial="$value" ;;
                        BRAND) brand="$value" ;;
                        SIZE) size="$value" ;;
                    esac
                done < "$METADATA_DIR/$device.meta"
                
                # Recreate gadget if we have all required data
                if [[ -n "$vid_pid" && -n "$vendor" && -n "$product" && -n "$serial" && -n "$brand" && -n "$size" ]]; then
                    # Create gadget structure
                    local gadget_path="$GADGET_ROOT/virtusb-$device"
                    mkdir -p "$gadget_path"/{strings/0x409,configs/c.1/strings/0x409,functions/mass_storage.0/lun.0}
                    
                    # Set vendor/product IDs
                    echo "0x${vid_pid%:*}" > "$gadget_path/idVendor"
                    echo "0x${vid_pid#*:}" > "$gadget_path/idProduct"
                    
                    # Set strings
                    echo "$vendor" > "$gadget_path/strings/0x409/manufacturer"
                    echo "$product" > "$gadget_path/strings/0x409/product"
                    echo "Config 1" > "$gadget_path/configs/c.1/strings/0x409/configuration"
                    
                    # Set mass storage
                    echo "$IMAGE_DIR/$device.img" > "$gadget_path/functions/mass_storage.0/lun.0/file"
                    
                    # Link function
                    ln -sf "$gadget_path/functions/mass_storage.0" "$gadget_path/configs/c.1/"
                    
                    # Try to enable the recreated gadget
                    enable_gadget "$device" "true" 2>/dev/null || {
                        log_error "Échec de l'activation du gadget recréé: $device"
                        remove_enabled_state "$device"
                        ((failed_count++))
                    }
                    ((restored_count++))
                else
                    log_error "Métadonnées incomplètes pour le gadget: $device"
                    remove_enabled_state "$device"
                    ((failed_count++))
                fi
            else
                # Missing metadata or image, remove from enabled state
                remove_enabled_state "$device"
                ((failed_count++))
            fi
        fi
    done
    
    # Clean up orphaned metadata files (devices that exist in metadata but not in configfs and not in enabled list)
    local orphaned_count=0
    for meta_file in "$METADATA_DIR"/*.meta; do
        [[ -f "$meta_file" ]] || continue
        
        # Extract name from metadata file
        local name=""
        while IFS='=' read -r key value; do
            value="${value//\"/}"
            [[ "$key" == "NAME" ]] && name="$value"
        done < "$meta_file"
        
        if [[ -n "$name" ]]; then
            # Check if gadget exists in configfs and is not in enabled list
            if [[ ! -d "$GADGET_ROOT/virtusb-$name" ]] && [[ ! " ${enabled_devices[*]} " =~ " ${name} " ]]; then
                rm -f "$meta_file"
                rm -f "$IMAGE_DIR/$name.img"
                remove_enabled_state "$name"
                ((orphaned_count++))
            fi
        fi
    done
    
    # Clean up orphaned gadgets (gadgets in configfs without metadata)
    local orphaned_gadgets=0
    for gadget_dir in "$GADGET_ROOT"/virtusb-*; do
        [[ -d "$gadget_dir" ]] || continue
        
        local gadget_name
        gadget_name=$(basename "$gadget_dir" | sed 's/^virtusb-//')
        
        if [[ ! -f "$METADATA_DIR/$gadget_name.meta" ]]; then
            # Remove orphaned gadget from configfs
            echo "" > "$gadget_dir/UDC" 2>/dev/null || true
            rm -f "$gadget_dir/configs/c.1/mass_storage.0" 2>/dev/null || true
            rmdir "$gadget_dir/configs/c.1/strings/0x409" 2>/dev/null || true
            rmdir "$gadget_dir/configs/c.1" 2>/dev/null || true
            rmdir "$gadget_dir/functions/mass_storage.0/lun.0" 2>/dev/null || true
            rmdir "$gadget_dir/functions/mass_storage.0" 2>/dev/null || true
            rmdir "$gadget_dir/functions" 2>/dev/null || true
            rmdir "$gadget_dir/strings/0x409" 2>/dev/null || true
            rmdir "$gadget_dir/strings" 2>/dev/null || true
            rmdir "$gadget_dir/os_desc" 2>/dev/null || true
            rmdir "$gadget_dir/webusb" 2>/dev/null || true
            rmdir "$gadget_dir/configs" 2>/dev/null || true
            rmdir "$gadget_dir" 2>/dev/null || true
            remove_enabled_state "$gadget_name"
            ((orphaned_gadgets++))
        fi
    done
    
    # Clean up orphaned image files (images without metadata)
    local orphaned_images=0
    for img_file in "$IMAGE_DIR"/*.img; do
        [[ -f "$img_file" ]] || continue
        
        local img_name
        img_name=$(basename "$img_file" .img)
        
        if [[ ! -f "$METADATA_DIR/$img_name.meta" ]]; then
            rm -f "$img_file"
            remove_enabled_state "$img_name"
            ((orphaned_images++))
        fi
    done
    
    # Log results
    if [[ $restored_count -gt 0 ]]; then
        log_success "Restauration après reboot: $restored_count gadgets recréés et activés"
    fi
    if [[ $failed_count -gt 0 ]]; then
        log_error "Échecs de restauration: $failed_count gadgets"
    fi
    if [[ $orphaned_count -gt 0 || $orphaned_gadgets -gt 0 || $orphaned_images -gt 0 ]]; then
        log_info "Nettoyage automatique après reboot: $orphaned_count métadonnées, $orphaned_gadgets gadgets, $orphaned_images images supprimés"
    fi
}

# Check root privileges
check_root() {
    [[ $EUID -eq 0 ]] || { log_error "Root privileges required"; exit 1; }
}

# Check and load kernel modules
check_modules() {
    local modules=("libcomposite" "usb_f_mass_storage")
    
    # Load basic modules
    for module in "${modules[@]}"; do
        lsmod | grep -q "^${module} " || {
            modprobe "$module" 2>/dev/null || {
                log_error "Unable to load module $module"
                exit 1
            }
        }
    done
    
    # Load dummy_hcd with multiple UDC instances (up to 30 UDCs supported)
    if ! lsmod | grep -q "^dummy_hcd "; then
        modprobe dummy_hcd num=30 2>/dev/null || {
            log_error "Unable to load dummy_hcd module"
            exit 1
        }
    fi
    
    # Verify UDC availability (simplified)
    if [[ ! -d /sys/class/udc ]] || [[ -z "$(ls /sys/class/udc/ 2>/dev/null || echo "")" ]]; then
        log_error "No UDC available after loading dummy_hcd"
        exit 1
    fi
}

# Check if configfs is mounted
check_configfs() {
    [[ -d "$GADGET_ROOT" ]] || {
        log_error "configfs not mounted"
        exit 1
    }
}

# Detect VM manager and provide usage hints
detect_vm_manager() {
    if [[ -f /etc/pve/version ]]; then
        echo "📋 Proxmox detected - Use: qm set <VM_ID> -usb0 host=<VID>:<PID>"
    elif [[ -f /etc/vmware-release ]] || [[ -d /vmfs ]]; then
        echo "📋 VMware ESXi detected - Use: vim-cmd vmsvc/device.connectusb <VM_ID> <VID>:<PID>"
    elif command -v virsh >/dev/null 2>&1; then
        echo "📋 KVM/QEMU detected - Use: virsh attach-device <VM_NAME> <XML_FILE>"
    elif command -v VBoxManage >/dev/null 2>&1; then
        echo "📋 VirtualBox detected - Use: VBoxManage controlvm <VM_NAME> usbattach <VID>:<PID>"
    else
        echo "📋 Generic Linux detected - Check your VM manager documentation"
    fi
}

# Check available disk space
check_disk_space() {
    local required_mb="$1"
    local available_mb
    
    available_mb=$(df -m "$IMAGE_DIR" | awk 'NR==2 {print $4}')
    [[ "$available_mb" -ge "$required_mb" ]] || {
        log_error "Insufficient disk space: ${available_mb}MB available, ${required_mb}MB required"
        exit 1
    }
}

# Convert size to MB
convert_size_to_mb() {
    local size="$1"
    
    case "$size" in
        *G) echo $(( ${size%G} * 1024 )) ;;
        *M) echo "${size%M}" ;;
        *K) echo $(( ${size%K} / 1024 )) ;;
        *)  echo "$size" ;;
    esac
}

# Validation

validate_name() {
    [[ "$1" =~ ^[a-zA-Z0-9_-]+$ ]] || { 
        log_error "Invalid name: $1 (use only letters, numbers, hyphens and underscores)"
        exit 1
    }
    [[ ${#1} -le 50 ]] || {
        log_error "Name too long: $1 (max 50 characters)"
        exit 1
    }
}

validate_size() {
    [[ "$1" =~ ^[0-9]+[KMG]?$ ]] || { 
        log_error "Invalid size: $1 (format: 1G, 512M, 8G, etc.)"
        exit 1
    }
    local size_mb=$(convert_size_to_mb "$1")
    [[ "$size_mb" -gt 0 && "$size_mb" -le 1048576 ]] || {
        log_error "Size out of range: $1 (1K to 1TB)"
        exit 1
    }
}

validate_brand() {
    case "$1" in
        sandisk|kingston|samsung|toshiba|lexar|pny|verbatim|transcend|adata|corsair) return 0 ;;
        *) 
            log_error "Invalid brand: $1"
            log_error "Supported brands: sandisk, kingston, samsung, toshiba, lexar, pny, verbatim, transcend, adata, corsair"
            exit 1
            ;;
    esac
}

# Brand management

# Get vendor/product information
get_vendor_product() {
    case "$1" in
        sandisk)    echo "SanDisk Corp." "Cruzer Blade" ;;
        kingston)   echo "Kingston Technology" "DataTraveler" ;;
        samsung)    echo "Samsung Electronics" "USB Flash Drive" ;;
        toshiba)    echo "Toshiba Corp." "TransMemory" ;;
        lexar)      echo "Lexar Media" "JumpDrive" ;;
        pny)        echo "PNY Technologies" "Attaché" ;;
        verbatim)   echo "Verbatim Corp." "Store 'n' Go" ;;
        transcend)  echo "Transcend Information" "JetFlash" ;;
        adata)      echo "ADATA Technology" "USB Flash Drive" ;;
        corsair)    echo "Corsair Memory" "Voyager" ;;
        *)          log_error "Unknown brand: $1"; exit 1 ;;
    esac
}

# Get VID:PID
get_vid_pid() {
    case "$1" in
        sandisk)    echo "0781:5567" ;;
        kingston)   echo "0951:1666" ;;
        samsung)    echo "04e8:61f5" ;;
        toshiba)    echo "0930:6545" ;;
        lexar)      echo "05dc:a4a5" ;;
        pny)        echo "154b:6545" ;;
        verbatim)   echo "18a5:0302" ;;
        transcend)  echo "0c76:0005" ;;
        adata)      echo "125f:c90a" ;;
        corsair)    echo "1b1c:1a0b" ;;
        *)          log_error "Unknown brand: $1"; exit 1 ;;
    esac
}

# Generate unique serial number
generate_serial() {
    printf "%08X%04X" "$(date +%s)" "$((RANDOM % 10000))"
}

# Image management

create_image() {
    local name="$1" size="$2"
    local image_path="$IMAGE_DIR/$name.img"
    local size_mb
    
    [[ -f "$image_path" ]] && { 
        log_error "Image already exists: $name"
        exit 1
    }
    
    size_mb=$(convert_size_to_mb "$size")
    check_disk_space "$size_mb"
    
    dd if=/dev/zero of="$image_path" bs=1M count="$size_mb" 2>/dev/null || {
        log_error "Failed to create image"
        exit 1
    }
    

}

# Gadget management

create_gadget() {
    local name="$1" vid_pid="$2" vendor="$3" product="$4" serial="$5" brand="$6" size="$7"
    local gadget_path="$GADGET_ROOT/virtusb-$name"
    
    [[ -d "$gadget_path" ]] && { 
        log_error "Gadget already exists: $name"
        exit 1
    }
    
    # Create gadget structure
    mkdir -p "$gadget_path"/{strings/0x409,configs/c.1/strings/0x409,functions/mass_storage.0/lun.0}
    
    # Set vendor/product IDs
    echo "0x${vid_pid%:*}" > "$gadget_path/idVendor"
    echo "0x${vid_pid#*:}" > "$gadget_path/idProduct"
    
    # Set strings
    echo "$vendor" > "$gadget_path/strings/0x409/manufacturer"
    echo "$product" > "$gadget_path/strings/0x409/product"
    echo "Config 1" > "$gadget_path/configs/c.1/strings/0x409/configuration"
    
    # Set mass storage
    echo "$IMAGE_DIR/$name.img" > "$gadget_path/functions/mass_storage.0/lun.0/file"
    
    # Link function
    ln -sf "$gadget_path/functions/mass_storage.0" "$gadget_path/configs/c.1/"
    
    # Save metadata
    cat > "$METADATA_DIR/$name.meta" << EOF
NAME="$name"
VID_PID="$vid_pid"
VENDOR="$vendor"
PRODUCT="$product"
SERIAL="$serial"
BRAND="$brand"
SIZE="$size"
CREATED_AT="$(date -Iseconds)"
EOF
    
    echo "🔧 Device $name is ready"
}

enable_gadget() {
    local name="$1"
    local silent="${2:-false}"
    local gadget_path="$GADGET_ROOT/virtusb-$name"
    
    [[ -d "$gadget_path" ]] || { 
        log_error "Gadget not found: $name"
        exit 1
    }
    
    # Check if already enabled
    if [[ -f "$gadget_path/UDC" ]]; then
        local current_udc
        current_udc=$(cat "$gadget_path/UDC" 2>/dev/null || echo "")
        if [[ -n "$current_udc" ]]; then
            # Only show message if not silent
            if [[ "$silent" != "true" ]]; then
                echo "🟢 Device $name already mounted"
            fi
            return 0
        fi
    fi
    
    # Find an available UDC
    local udc_list=()
    for udc_dir in /sys/class/udc/*; do
        [[ -d "$udc_dir" ]] && udc_list+=("$(basename "$udc_dir")")
    done
    
    [[ ${#udc_list[@]} -eq 0 ]] && {
        log_error "No UDC available"
        exit 1
    }
    
    # Find an UDC that's not currently in use
    for udc in "${udc_list[@]}"; do
        local in_use=false
        
        # Check if this UDC is already assigned to another gadget
        for gadget_dir in "$GADGET_ROOT"/virtusb-*; do
            [[ -d "$gadget_dir" ]] || continue
            [[ -f "$gadget_dir/UDC" ]] || continue
            
            local gadget_udc
            gadget_udc=$(cat "$gadget_dir/UDC" 2>/dev/null || echo "")
            if [[ "$gadget_udc" == "$udc" ]]; then
                in_use=true
                break
            fi
        done
        
        if [[ "$in_use" == "false" ]]; then
            # Assign this UDC to the gadget
            echo "$udc" > "$gadget_path/UDC" 2>/dev/null || {
                log_error "Failed to assign UDC $udc to gadget '$name'"
                exit 1
            }
            
            # Verify the device is actually detected
            local vid_pid=""
            while IFS='=' read -r key value; do
                value="${value//\"/}"
                [[ "$key" == "VID_PID" ]] && vid_pid="$value"
            done < "$METADATA_DIR/$name.meta"
            
            # Wait a moment for the device to be detected
            sleep 1
            
            # Check if device appears in lsusb
            if lsusb | grep -q "$vid_pid"; then
                # Only show message if not silent
                if [[ "$silent" != "true" ]]; then
                    echo "🟢 Device $name mounted"
                fi
                save_enabled_state "$name"
                return 0
            else
                # Device not detected, clean up
                echo "" > "$gadget_path/UDC" 2>/dev/null || true
                log_error "Device $name failed to mount - not detected by system"
                exit 1
            fi
        fi
    done
    
    log_error "No available UDC for gadget '$name'"
    exit 1
}

disable_gadget() {
    local name="$1"
    local gadget_path="$GADGET_ROOT/virtusb-$name"
    
    [[ -d "$gadget_path" ]] || { 
        log_error "Gadget not found: $name"
        exit 1
    }
    
    echo "" > "$gadget_path/UDC" 2>/dev/null || true
    remove_enabled_state "$name"
    echo "⏹️ Device $name unmounted"
}

delete_gadget() {
    local name="$1"
    local gadget_path="$GADGET_ROOT/virtusb-$name"
    local gadget_exists=false
    local files_removed=false
    
    # Check if gadget exists in configfs
    if [[ -d "$gadget_path" ]]; then
        gadget_exists=true
        
        # Disable first (silently)
        [[ -f "$gadget_path/UDC" && -s "$gadget_path/UDC" ]] && echo "" > "$gadget_path/UDC" 2>/dev/null || true
        
        # Remove gadget structure (configfs requires specific order)
        rm -f "$gadget_path/configs/c.1/mass_storage.0" 2>/dev/null || true
        rmdir "$gadget_path/configs/c.1/strings/0x409" 2>/dev/null || true
        rmdir "$gadget_path/configs/c.1" 2>/dev/null || true
        rmdir "$gadget_path/functions/mass_storage.0/lun.0" 2>/dev/null || true
        rmdir "$gadget_path/functions/mass_storage.0" 2>/dev/null || true
        rmdir "$gadget_path/functions" 2>/dev/null || true
        rmdir "$gadget_path/strings/0x409" 2>/dev/null || true
        rmdir "$gadget_path/strings" 2>/dev/null || true
        rmdir "$gadget_path/os_desc" 2>/dev/null || true
        rmdir "$gadget_path/webusb" 2>/dev/null || true
        rmdir "$gadget_path/configs" 2>/dev/null || true
        
        # Clear configfs files before removal
        echo "" > "$gadget_path/UDC" 2>/dev/null || true
        echo "" > "$gadget_path/idVendor" 2>/dev/null || true
        echo "" > "$gadget_path/idProduct" 2>/dev/null || true
        echo "" > "$gadget_path/bcdDevice" 2>/dev/null || true
        echo "" > "$gadget_path/bcdUSB" 2>/dev/null || true
        echo "" > "$gadget_path/bDeviceClass" 2>/dev/null || true
        echo "" > "$gadget_path/bDeviceProtocol" 2>/dev/null || true
        echo "" > "$gadget_path/bDeviceSubClass" 2>/dev/null || true
        echo "" > "$gadget_path/bMaxPacketSize0" 2>/dev/null || true
        echo "" > "$gadget_path/max_speed" 2>/dev/null || true
        
        # Force removal of gadget directory
        rmdir "$gadget_path" 2>/dev/null || {
            log_error "Failed to remove gadget directory for '$name', attempting module reload..."
            # If removal fails, try to force it by reloading the module
            modprobe -r dummy_hcd 2>/dev/null || true
            sleep 2
            modprobe dummy_hcd num=30 2>/dev/null || {
                log_error "Failed to reload dummy_hcd module"
                return 1
            }
            # Try removal again after module reload
            rmdir "$gadget_path" 2>/dev/null || {
                log_error "Still failed to remove gadget directory for '$name' after module reload"
                return 1
            }
        }
    fi
    
    # Remove data files (metadata and image)
    if [[ -f "$METADATA_DIR/$name.meta" ]]; then
        rm -f "$METADATA_DIR/$name.meta"
        files_removed=true
    fi
    
    if [[ -f "$IMAGE_DIR/$name.img" ]]; then
        rm -f "$IMAGE_DIR/$name.img"
        files_removed=true
    fi
    
    # Remove from enabled state
    remove_enabled_state "$name"
    
    # Provide appropriate feedback
    if [[ "$gadget_exists" == "true" ]]; then
        echo "🗑️ Device $name removed"
    elif [[ "$files_removed" == "true" ]]; then
        echo "🧹 Orphaned files for device $name cleaned up"
    else
        log_error "Gadget not found: $name"
        return 1
    fi
    
    return 0
}

# Integrity checking
check_gadget_integrity() {
    local name="$1"
    [[ -f "$METADATA_DIR/$name.meta" ]] && \
    [[ -d "$GADGET_ROOT/virtusb-$name" ]] && \
    [[ -f "$IMAGE_DIR/$name.img" ]]
}

# System integrity check and repair
check_system_integrity() {
    # Check for orphaned metadata files
    local orphaned_metadata=0
    for meta_file in "$METADATA_DIR"/*.meta; do
        [[ -f "$meta_file" ]] || continue
        
        local name=""
        while IFS='=' read -r key value; do
            value="${value//\"/}"
            [[ "$key" == "NAME" ]] && name="$value"
        done < "$meta_file"
        
        if [[ -n "$name" && ! -d "$GADGET_ROOT/virtusb-$name" ]]; then
            ((orphaned_metadata++))
        fi
    done
    
    # Check for orphaned gadgets (gadgets without metadata)
    local orphaned_gadgets=0
    for gadget_dir in "$GADGET_ROOT"/virtusb-*; do
        [[ -d "$gadget_dir" ]] || continue
        
        local gadget_name
        gadget_name=$(basename "$gadget_dir" | sed 's/^virtusb-//')
        
        if [[ ! -f "$METADATA_DIR/$gadget_name.meta" ]]; then
            ((orphaned_gadgets++))
        fi
    done
    
    if [[ $orphaned_metadata -gt 0 || $orphaned_gadgets -gt 0 ]]; then
        return 1
    else
        return 0
    fi
}

# Cleanup

purge() {
    echo "🧹 Nettoyage automatique des fichiers orphelins..."
    
    local orphaned_metadata=0
    local orphaned_gadgets=0
    local orphaned_images=0
    
    # Phase 1: Clean up orphaned metadata files (devices that exist in metadata but not in configfs)
    for meta_file in "$METADATA_DIR"/*.meta; do
        [[ -f "$meta_file" ]] || continue
        
        # Extract name from metadata file
        local name=""
        while IFS='=' read -r key value; do
            value="${value//\"/}"
            [[ "$key" == "NAME" ]] && name="$value"
        done < "$meta_file"
        
        if [[ -n "$name" ]]; then
            # Check if gadget exists in configfs
            if [[ ! -d "$GADGET_ROOT/virtusb-$name" ]]; then
                rm -f "$meta_file"
                rm -f "$IMAGE_DIR/$name.img"
                remove_enabled_state "$name"
                ((orphaned_metadata++))
            fi
        fi
    done
    
    # Phase 2: Clean up orphaned gadgets (gadgets in configfs without metadata)
    for gadget_dir in "$GADGET_ROOT"/virtusb-*; do
        [[ -d "$gadget_dir" ]] || continue
        
        local gadget_name
        gadget_name=$(basename "$gadget_dir" | sed 's/^virtusb-//')
        
        if [[ ! -f "$METADATA_DIR/$gadget_name.meta" ]]; then
            # Remove orphaned gadget from configfs
            echo "" > "$gadget_dir/UDC" 2>/dev/null || true
            rm -f "$gadget_dir/configs/c.1/mass_storage.0" 2>/dev/null || true
            rmdir "$gadget_dir/configs/c.1/strings/0x409" 2>/dev/null || true
            rmdir "$gadget_dir/configs/c.1" 2>/dev/null || true
            rmdir "$gadget_dir/functions/mass_storage.0/lun.0" 2>/dev/null || true
            rmdir "$gadget_dir/functions/mass_storage.0" 2>/dev/null || true
            rmdir "$gadget_dir/functions" 2>/dev/null || true
            rmdir "$gadget_dir/strings/0x409" 2>/dev/null || true
            rmdir "$gadget_dir/strings" 2>/dev/null || true
            rmdir "$gadget_dir/os_desc" 2>/dev/null || true
            rmdir "$gadget_dir/webusb" 2>/dev/null || true
            rmdir "$gadget_dir/configs" 2>/dev/null || true
            rmdir "$gadget_dir" 2>/dev/null || true
            remove_enabled_state "$gadget_name"
            ((orphaned_gadgets++))
        fi
    done
    
    # Phase 3: Clean up orphaned image files (images without metadata)
    for img_file in "$IMAGE_DIR"/*.img; do
        [[ -f "$img_file" ]] || continue
        
        local img_name
        img_name=$(basename "$img_file" .img)
        
        if [[ ! -f "$METADATA_DIR/$img_name.meta" ]]; then
            rm -f "$img_file"
            remove_enabled_state "$img_name"
            ((orphaned_images++))
        fi
    done
    
    if [[ $orphaned_metadata -gt 0 || $orphaned_gadgets -gt 0 || $orphaned_images -gt 0 ]]; then
        echo "✅ Nettoyage automatique terminé: $orphaned_metadata métadonnées, $orphaned_gadgets gadgets, $orphaned_images images supprimés"
    else
        echo "✅ Aucun fichier orphelin trouvé"
    fi
    
    # Ask for confirmation to remove ALL remaining gadgets
    echo ""
    echo "Supprimer TOUS les gadgets restants ? (y/n)"
    read -r response
    
    if [[ "$response" != "y" ]]; then
        echo "❌ Opération annulée"
        return
    fi
    
    echo "🧹 Suppression de tous les gadgets..."
    
    # Phase 4: Remove all gadgets with metadata
    local removed_count=0
    for meta_file in "$METADATA_DIR"/*.meta; do
        [[ -f "$meta_file" ]] || continue
        
        # Extract name from metadata file
        local name=""
        while IFS='=' read -r key value; do
            value="${value//\"/}"
            [[ "$key" == "NAME" ]] && name="$value"
        done < "$meta_file"
        
        if [[ -n "$name" ]]; then
            echo "🗑️ Suppression du gadget avec métadonnées: $name"
            delete_gadget "$name" 2>/dev/null || {
                log_error "Échec de la suppression du gadget: $name"
            }
            ((removed_count++))
        fi
    done
    
    echo "✅ $removed_count gadgets avec métadonnées supprimés"
    
    # Phase 5: Remove all remaining orphaned gadgets (double check)
    local orphaned_count=0
    for gadget_dir in "$GADGET_ROOT"/virtusb-*; do
        [[ -d "$gadget_dir" ]] || continue
        
        local gadget_name
        gadget_name=$(basename "$gadget_dir" | sed 's/^virtusb-//')
        
        echo "🗑️ Suppression du gadget orphelin: $gadget_name"
        
        # Disable first
        [[ -f "$gadget_dir/UDC" && -s "$gadget_dir/UDC" ]] && echo "" > "$gadget_dir/UDC" 2>/dev/null || true
        
        # Remove structure (configfs requires specific order)
        rm -f "$gadget_dir/configs/c.1/mass_storage.0" 2>/dev/null || true
        rmdir "$gadget_dir/configs/c.1/strings/0x409" 2>/dev/null || true
        rmdir "$gadget_dir/configs/c.1" 2>/dev/null || true
        rmdir "$gadget_dir/functions/mass_storage.0/lun.0" 2>/dev/null || true
        rmdir "$gadget_dir/functions/mass_storage.0" 2>/dev/null || true
        rmdir "$gadget_dir/functions" 2>/dev/null || true
        rmdir "$gadget_dir/strings/0x409" 2>/dev/null || true
        rmdir "$gadget_dir/strings" 2>/dev/null || true
        rmdir "$gadget_dir/os_desc" 2>/dev/null || true
        rmdir "$gadget_dir/webusb" 2>/dev/null || true
        rmdir "$gadget_dir/configs" 2>/dev/null || true
        
        # Clear configfs files before removal
        echo "" > "$gadget_dir/UDC" 2>/dev/null || true
        echo "" > "$gadget_dir/idVendor" 2>/dev/null || true
        echo "" > "$gadget_dir/idProduct" 2>/dev/null || true
        echo "" > "$gadget_dir/bcdDevice" 2>/dev/null || true
        echo "" > "$gadget_dir/bcdUSB" 2>/dev/null || true
        echo "" > "$gadget_dir/bDeviceClass" 2>/dev/null || true
        echo "" > "$gadget_dir/bDeviceProtocol" 2>/dev/null || true
        echo "" > "$gadget_dir/bDeviceSubClass" 2>/dev/null || true
        echo "" > "$gadget_dir/bMaxPacketSize0" 2>/dev/null || true
        echo "" > "$gadget_dir/max_speed" 2>/dev/null || true
        
        # Force removal of gadget directory
        rmdir "$gadget_dir" 2>/dev/null || {
            log_error "Échec de la suppression du répertoire gadget orphelin: $gadget_dir"
        }
        ((orphaned_count++))
    done
    
    echo "✅ $orphaned_count gadgets orphelins supprimés"
    
    # Phase 6: Clean data directories
    echo "🧹 Nettoyage des répertoires de données..."
    rm -rf "$IMAGE_DIR"/* 2>/dev/null || true
    rm -rf "$METADATA_DIR"/* 2>/dev/null || true
    rm -f "$ENABLED_STATE_FILE" 2>/dev/null || true
    
    # Recreate empty directories
    mkdir -p "$IMAGE_DIR"
    mkdir -p "$METADATA_DIR"
    
    echo "🧹 Tous les gadgets ont été nettoyés"
}

# Display

list_gadgets() {
    echo "NAME                 ENABLED  VID:PID      BRAND         SERIAL"
    echo "---------------------------------------------------------------"
    
    for meta_file in "$METADATA_DIR"/*.meta; do
        [[ -f "$meta_file" ]] || continue
        
        # Read metadata safely
        local name=""
        local vid_pid=""
        local brand=""
        local serial=""
        
        while IFS='=' read -r key value; do
            # Remove quotes from value
            value="${value//\"/}"
            case "$key" in
                NAME) name="$value" ;;
                VID_PID) vid_pid="$value" ;;
                BRAND) brand="$value" ;;
                SERIAL) serial="$value" ;;
            esac
        done < "$meta_file"
        
        [[ -n "$name" ]] || continue
        
        local gadget_path="$GADGET_ROOT/virtusb-$name"
        local status="❌"
        
        # Check if gadget exists physically
        if [[ ! -d "$gadget_path" ]]; then
            status="🔴"
            name="$name (orphelin)"
        elif [[ -f "$gadget_path/UDC" ]]; then
            local udc_content
            udc_content=$(cat "$gadget_path/UDC" 2>/dev/null || echo "")
            if [[ -n "$udc_content" ]]; then
                # Additional check: verify device appears in lsusb
                if lsusb | grep -q "$vid_pid"; then
                    status="✅"
                fi
            fi
        fi
        
        printf "%-20s %s  %-10s    %-12s %s\n" "$name" "$status" "${vid_pid:-}" "${brand:-}" "${serial:-}"
    done
}

show_help() {
    cat << EOF
$SCRIPT_NAME - Virtual USB Device Manager
Version: $SCRIPT_VERSION - Universal Linux Support

Usage: $SCRIPT_NAME <command> [options]

Commands:
  create <name> --size <size> --brand <brand>  Create a virtual USB device
  enable <name>                                Enable/mount a device
  disable <name>                               Disable/unmount a device
  delete <name>                                Delete a device
  list                                         List all devices
  purge                                        Remove ALL devices (includes orphan cleanup)
  help                                         Show this help

Options:
  --size <size>    Device size (e.g., 1G, 512M, 8G)
  --brand <brand>  Device brand

Supported brands:
  sandisk, kingston, samsung, toshiba, lexar, pny, verbatim, transcend, adata, corsair

Examples:
  sudo $SCRIPT_NAME create myusb --size 8G --brand sandisk
  sudo $SCRIPT_NAME enable myusb
  sudo $SCRIPT_NAME list
  sudo $SCRIPT_NAME delete myusb

$(detect_vm_manager)

Notes:
  • Root privileges required for all operations
  • Devices are automatically restored after system reboot
  • Use 'virtusb purge' to clean up orphaned files
  • Devices appear as real USB storage devices to the system

EOF
}

# Main

main() {
    check_root
    
    case "${1:-}" in
        create|enable|disable|delete|list|purge|help|--help|-h)
            check_modules
            check_configfs
            restore_enabled_devices
            ;;
        *)
            log_error "Unknown command: ${1:-}"
            show_help
            exit 1
            ;;
    esac
    
    case "${1:-}" in
        create)
            [[ $# -lt 4 ]] && { 
                log_error "Usage: $SCRIPT_NAME create <name> --size <size> --brand <brand>"
                exit 1
            }
            
            local name="$2" size brand
            
            # Parse arguments
            local i=3
            while [[ $i -le $# ]]; do
                local current_arg="${!i}"
                case "$current_arg" in
                    --size)
                        [[ $((i+1)) -le $# ]] || { log_error "Size value required after --size"; exit 1; }
                        local next_i=$((i+1))
                        size="${!next_i}"
                        i=$((i+2))
                        ;;
                    --brand)
                        [[ $((i+1)) -le $# ]] || { log_error "Brand value required after --brand"; exit 1; }
                        local next_i=$((i+1))
                        brand="${!next_i}"
                        i=$((i+2))
                        ;;
                    *)
                        log_error "Unknown option: $current_arg"
                        exit 1
                        ;;
                esac
            done
            
            [[ -z "${size:-}" ]] && { log_error "Size required (--size)"; exit 1; }
            [[ -z "${brand:-}" ]] && { log_error "Brand required (--brand)"; exit 1; }
            
            validate_name "$name"
            validate_size "$size"
            validate_brand "$brand"
            
            # Create device
            local vid_pid vendor product serial
            vid_pid=$(get_vid_pid "$brand")
            read -r vendor product <<< "$(get_vendor_product "$brand")"
            serial=$(generate_serial)
            
            create_image "$name" "$size"
            create_gadget "$name" "$vid_pid" "$vendor" "$product" "$serial" "$brand" "$size"
            ;;
        enable)
            [[ $# -ne 2 ]] && { log_error "Usage: $SCRIPT_NAME enable <name>"; exit 1; }
            validate_name "$2"
            enable_gadget "$2"
            ;;
        disable)
            [[ $# -ne 2 ]] && { log_error "Usage: $SCRIPT_NAME disable <name>"; exit 1; }
            validate_name "$2"
            disable_gadget "$2"
            ;;
        delete)
            [[ $# -ne 2 ]] && { log_error "Usage: $SCRIPT_NAME delete <name>"; exit 1; }
            validate_name "$2"
            delete_gadget "$2"
            ;;
        list)
            list_gadgets
            ;;
        purge)
            purge
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "Unknown command: ${1:-}"
            show_help
            exit 1
            ;;
    esac
}

# Execute

main "$@"