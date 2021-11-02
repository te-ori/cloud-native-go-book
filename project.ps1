$mainGitHubUrl="https://github.com/cloud-native-go/examples"
Write-Host "PROJECT helpers loaded"

$keyValueStoreBaseUrl = "https://127.0.0.1:8080/v1/"
function Open-Github {
    $browserPath = GET-DefaultBrowserPath
    
    Start-Process $browserPath -ArgumentList $mainGitHubUrl
}

function Generate-RandomString {
    param(
        [Parameter(Position=1)]
        [string]
        $len = 5
    )

    -join (97..122 | Get-Random -Count $len | % {[char]$_})
}

function Put-Key {
    [CmdletBinding()]
    param (
        [Parameter(Mandatory = 1, Position=1)]
        [string]
        $key,
        [Parameter(Mandatory = 1, Position=2)]
        [string]
        $value
    )

    curl -X PUT -d "$value" "$keyValueStoreBaseUrl$key"
}

function Put-RandomKey {
    $key = (Generate-RandomString) 
    $value = (Generate-RandomString 15)

    Put-Key $key $value
    
    Write-Host "KEY       VALUE"
    Write-Host "-----------------------------"
    Write-Host "$key     $value"
}

function Get-Key {
    [CmdletBinding()]
    param (
        [Parameter(Mandatory = 1, Position=1)]
        [string]
        $key
    )

    curl "$keyValueStoreBaseUrl$key"
}


function Delete-Key {
    [CmdletBinding()]
    param (
        [Parameter(Mandatory = 1, Position=1)]
        [string]
        $key
    )

    curl -X DELETE "$keyValueStoreBaseUrl$key"
}