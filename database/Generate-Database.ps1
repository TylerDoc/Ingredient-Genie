#Requires -Modules PSSQLite

Import-Module PSSQLite

$ErrorActionPreference = 'Stop'

function ConvertTo-TrimmedOrNull {
    param (
        [Parameter()]
        [AllowNull()]
        [object] $Value
    )

    if ($null -eq $Value) {
        return $null
    }

    $text = ([string]$Value).Trim()

    if ([string]::IsNullOrWhiteSpace($text)) {
        return $null
    }

    return $text
}

function New-PreparedSQLiteCommand {
    param (
        [Parameter(Mandatory)]
        $Connection,

        [Parameter(Mandatory)]
        $Transaction,

        [Parameter(Mandatory)]
        [string] $CommandText,

        [Parameter(Mandatory)]
        [hashtable] $ParameterTypes
    )

    $command = $Connection.CreateCommand()
    $command.Transaction = $Transaction
    $command.CommandText = $CommandText

    foreach ($name in $ParameterTypes.Keys) {
        $parameter = $command.CreateParameter()
        $parameter.ParameterName = "@$name"
        $parameter.DbType = $ParameterTypes[$name]

        [void]$command.Parameters.Add($parameter)
    }

    $command.Prepare()

    return $command
}

function Set-SQLiteCommandValues {
    param (
        [Parameter(Mandatory)]
        $Command,

        [Parameter(Mandatory)]
        [hashtable] $Values
    )

    foreach ($name in $Values.Keys) {
        $value = $Values[$name]

        $databaseValue = if ($null -eq $value) {
            [DBNull]::Value
        }
        else {
            $value
        }

        $Command.Parameters["@$name"].Value = $databaseValue
    }
}

$url = "https://www.themealdb.com/api/json/v1/1"

$meals = 'a'..'z' | ForEach-Object {
    (Invoke-RestMethod "$url/search.php?f=$_").meals
} | Where-Object { $null -ne $_ }

$normalizedMeals = foreach ($meal in $meals) {
    $ingredients = for ($i = 1; $i -le 20; $i++) {
        $ingredientProperty = "strIngredient$i"
        $measureProperty = "strMeasure$i"

        $ingredientName = ConvertTo-TrimmedOrNull -Value $meal.$ingredientProperty

        if ($null -eq $ingredientName) {
            continue
        }

        [pscustomobject]@{
            Position       = $i
            Name           = $ingredientName
            NormalizedName = (
                $ingredientName -replace '\s+', ' '
            ).ToLowerInvariant()
            MeasureText    = ConvertTo-TrimmedOrNull -Value $meal.$measureProperty
        }
    }

    [pscustomobject]@{
        ExternalMealId = ConvertTo-TrimmedOrNull -Value $meal.idMeal
        Name           = ConvertTo-TrimmedOrNull -Value $meal.strMeal
        AlternateName  = ConvertTo-TrimmedOrNull -Value $meal.strMealAlternate
        Category       = ConvertTo-TrimmedOrNull -Value $meal.strCategory
        Area           = ConvertTo-TrimmedOrNull -Value $meal.strArea
        Country        = ConvertTo-TrimmedOrNull -Value $meal.strCountry
        Instructions   = ConvertTo-TrimmedOrNull -Value $meal.strInstructions
        YoutubeUrl     = ConvertTo-TrimmedOrNull -Value $meal.strYoutube
        SourceUrl      = ConvertTo-TrimmedOrNull -Value $meal.strSource
        Ingredients    = @($ingredients)
    }
}

$databasePath = 'meals.sqlite'

if (Test-Path $databasePath) {
    Remove-Item $databasePath -Force
}

$connection = New-SQLiteConnection -DataSource $databasePath
$transaction = $null
$commands = @()

try {
    $schema = @'
PRAGMA foreign_keys = ON;

CREATE TABLE Meal (
    MealId          INTEGER PRIMARY KEY,
    ExternalMealId  TEXT UNIQUE,
    Name            TEXT NOT NULL,
    AlternateName   TEXT,
    Category        TEXT,
    Area            TEXT,
    Country         TEXT,
    Instructions    TEXT,
    YoutubeUrl      TEXT,
    SourceUrl       TEXT
);

CREATE TABLE Ingredient (
    IngredientId    INTEGER PRIMARY KEY,
    Name            TEXT NOT NULL,
    NormalizedName  TEXT NOT NULL COLLATE NOCASE UNIQUE
);

CREATE TABLE MealIngredient (
    MealId          INTEGER NOT NULL,
    IngredientId    INTEGER NOT NULL,
    Position        INTEGER NOT NULL CHECK (Position >= 1),
    MeasureText     TEXT,

    PRIMARY KEY (MealId, Position),

    FOREIGN KEY (MealId)
        REFERENCES Meal (MealId)
        ON DELETE CASCADE,

    FOREIGN KEY (IngredientId)
        REFERENCES Ingredient (IngredientId)
);

CREATE INDEX IX_MealIngredient_IngredientId
    ON MealIngredient (IngredientId);
'@

    $schemaQueryParams = @{
        SQLiteConnection = $connection
        Query            = $schema
        ErrorAction      = 'Stop'
    }

    Invoke-SqliteQuery @schemaQueryParams

    $transaction = $connection.BeginTransaction()

    $insertMealSql = @'
INSERT INTO Meal (
    ExternalMealId,
    Name,
    AlternateName,
    Category,
    Area,
    Country,
    Instructions,
    YoutubeUrl,
    SourceUrl
)
VALUES (
    @ExternalMealId,
    @Name,
    @AlternateName,
    @Category,
    @Area,
    @Country,
    @Instructions,
    @YoutubeUrl,
    @SourceUrl
);
'@

    $insertMealCommandParams = @{
        Connection  = $connection
        Transaction = $transaction
        CommandText = $insertMealSql
        ParameterTypes = @{
            ExternalMealId = [System.Data.DbType]::String
            Name           = [System.Data.DbType]::String
            AlternateName  = [System.Data.DbType]::String
            Category       = [System.Data.DbType]::String
            Area           = [System.Data.DbType]::String
            Country        = [System.Data.DbType]::String
            Instructions   = [System.Data.DbType]::String
            YoutubeUrl     = [System.Data.DbType]::String
            SourceUrl      = [System.Data.DbType]::String
        }
    }

    $insertMealCommand = New-PreparedSQLiteCommand @insertMealCommandParams
    $commands += $insertMealCommand

    $selectMealIdSql = @'
SELECT MealId
FROM Meal
WHERE ExternalMealId = @ExternalMealId;
'@

    $selectMealIdCommandParams = @{
        Connection  = $connection
        Transaction = $transaction
        CommandText = $selectMealIdSql
        ParameterTypes = @{
            ExternalMealId = [System.Data.DbType]::String
        }
    }

    $selectMealIdCommand = New-PreparedSQLiteCommand @selectMealIdCommandParams
    $commands += $selectMealIdCommand

    $insertIngredientSql = @'
INSERT OR IGNORE INTO Ingredient (
    Name,
    NormalizedName
)
VALUES (
    @Name,
    @NormalizedName
);
'@

    $insertIngredientCommandParams = @{
        Connection  = $connection
        Transaction = $transaction
        CommandText = $insertIngredientSql
        ParameterTypes = @{
            Name           = [System.Data.DbType]::String
            NormalizedName = [System.Data.DbType]::String
        }
    }

    $insertIngredientCommand = New-PreparedSQLiteCommand @insertIngredientCommandParams
    $commands += $insertIngredientCommand

    $selectIngredientIdSql = @'
SELECT IngredientId
FROM Ingredient
WHERE NormalizedName = @NormalizedName;
'@

    $selectIngredientIdCommandParams = @{
        Connection  = $connection
        Transaction = $transaction
        CommandText = $selectIngredientIdSql
        ParameterTypes = @{
            NormalizedName = [System.Data.DbType]::String
        }
    }

    $selectIngredientIdCommand = New-PreparedSQLiteCommand @selectIngredientIdCommandParams
    $commands += $selectIngredientIdCommand


    $insertMealIngredientSql = @'
INSERT INTO MealIngredient (
    MealId,
    IngredientId,
    Position,
    MeasureText
)
VALUES (
    @MealId,
    @IngredientId,
    @Position,
    @MeasureText
);
'@

    $insertMealIngredientCommandParams = @{
        Connection  = $connection
        Transaction = $transaction
        CommandText = $insertMealIngredientSql
        ParameterTypes = @{
            MealId       = [System.Data.DbType]::Int64
            IngredientId = [System.Data.DbType]::Int64
            Position     = [System.Data.DbType]::Int32
            MeasureText  = [System.Data.DbType]::String
        }
    }

    $insertMealIngredientCommand =
        New-PreparedSQLiteCommand @insertMealIngredientCommandParams

    $commands += $insertMealIngredientCommand

    foreach ($meal in $normalizedMeals) {
        if (
            [string]::IsNullOrWhiteSpace($meal.ExternalMealId) -or
            [string]::IsNullOrWhiteSpace($meal.Name)
        ) {
            Write-Warning 'Skipping a meal without an external ID or name.'
            continue
        }

        $mealValues = @{
            ExternalMealId = $meal.ExternalMealId
            Name           = $meal.Name
            AlternateName  = $meal.AlternateName
            Category       = $meal.Category
            Area           = $meal.Area
            Country        = $meal.Country
            Instructions   = $meal.Instructions
            YoutubeUrl     = $meal.YoutubeUrl
            SourceUrl      = $meal.SourceUrl
        }

        Set-SQLiteCommandValues `
            -Command $insertMealCommand `
            -Values $mealValues

        [void]$insertMealCommand.ExecuteNonQuery()

        Set-SQLiteCommandValues `
            -Command $selectMealIdCommand `
            -Values @{
                ExternalMealId = $meal.ExternalMealId
            }

        $mealId = $selectMealIdCommand.ExecuteScalar()

        if ($null -eq $mealId -or $mealId -is [DBNull]) {
            throw "Could not obtain the database ID for meal '$($meal.Name)'."
        }

        foreach ($ingredient in $meal.Ingredients) {
            Set-SQLiteCommandValues `
                -Command $insertIngredientCommand `
                -Values @{
                    Name           = $ingredient.Name
                    NormalizedName = $ingredient.NormalizedName
                }

            [void]$insertIngredientCommand.ExecuteNonQuery()

            Set-SQLiteCommandValues `
                -Command $selectIngredientIdCommand `
                -Values @{
                    NormalizedName = $ingredient.NormalizedName
                }

            $ingredientId = $selectIngredientIdCommand.ExecuteScalar()

            if ($null -eq $ingredientId -or $ingredientId -is [DBNull]) {
                throw "Could not obtain the ID for ingredient '$($ingredient.Name)'."
            }

            Set-SQLiteCommandValues `
                -Command $insertMealIngredientCommand `
                -Values @{
                    MealId       = [long]$mealId
                    IngredientId = [long]$ingredientId
                    Position     = [int]$ingredient.Position
                    MeasureText  = $ingredient.MeasureText
                }

            [void]$insertMealIngredientCommand.ExecuteNonQuery()
        }
    }

    $transaction.Commit()

    $summarySql = @'
SELECT
    (SELECT COUNT(*) FROM Meal) AS MealCount,
    (SELECT COUNT(*) FROM Ingredient) AS IngredientCount,
    (SELECT COUNT(*) FROM MealIngredient) AS MealIngredientCount;
'@

    $summaryQueryParams = @{
        SQLiteConnection = $connection
        Query            = $summarySql
        ErrorAction      = 'Stop'
    }

    $summary = Invoke-SqliteQuery @summaryQueryParams
    $summary | Format-List
}
catch {
    if ($null -ne $transaction) {
        try {
            $transaction.Rollback()
        }
        catch {
            Write-Warning "The transaction could not be rolled back: $_"
        }
    }

    throw
}
finally {
    foreach ($command in $commands) {
        if ($null -ne $command) {
            $command.Dispose()
        }
    }

    if ($null -ne $transaction) {
        $transaction.Dispose()
    }

    if ($null -ne $connection) {
        $connection.Close()
        $connection.Dispose()
    }
}