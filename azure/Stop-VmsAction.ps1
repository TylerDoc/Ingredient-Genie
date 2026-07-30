<#
    This script is used as part of an alert action to stop all
    VMs in the resource group if we hit ~90% of our budget
#>
param(
    [Parameter(Mandatory = $false)]
    [object] $WebhookData
)

$ErrorActionPreference = 'Stop'

Disable-AzContextAutosave -Scope Process

$SubscriptionId = '<id>'
$TargetResourceGroup = 'ingredient-genie'

Write-Output "authenticating with the automation account managed identity."

$context = (Connect-AzAccount -Identity).Context

$context = Set-AzContext -SubscriptionId $SubscriptionId -DefaultProfile $context

Write-Output "enumerating virtual machines in resource group '$TargetResourceGroup'."

$vms = Get-AzVM -ResourceGroupName $TargetResourceGroup -Status -DefaultProfile $context
if (-not $vms) {
    Write-Output "no virtual machines were found in '$TargetResourceGroup'."
    return
}

$failedVMs = [System.Collections.Generic.List[string]]::new()
foreach ($vm in $vms) {
    $powerState = (
        $vm.Statuses |
        Where-Object { $_.Code -like 'PowerState/*' } |
        Select-Object -First 1
    ).Code

    if ($powerState -in @(
        'PowerState/deallocated',
        'PowerState/deallocating'
    )) {
        Write-Output "skipping '$($vm.Name)'. Current state: $powerState"
        continue
    }

    try {
        Write-Output "deallocating VM '$($vm.Name)'. Current state: $powerState"
        Stop-AzVM -ResourceGroupName $TargetResourceGroup -Name $vm.Name -Force -DefaultProfile $context
        Write-Output "successfully deallocated '$($vm.Name)'."
    }
    catch {
        Write-Error "failed to deallocate '$($vm.Name)': $($_.Exception.Message)"
        $failedVMs.Add($vm.Name)
    }
}

if ($failedVMs.Count -gt 0) {
    throw "one or more VMs failed to deallocate: $($failedVMs -join ', ')"
}

Write-Output "all VMs in '$TargetResourceGroup' were deallocated."