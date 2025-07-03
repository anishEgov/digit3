## Introduction:

This library provides a flexible, human-readable ID generation system driven entirely by user-defined templates. It supports dynamic variables, formatted dates, customizable sequences (with scoped resets and padding), and random suffixes — all controlled via a simple JSON configuration. Postgres sequences are used under the hood to guarantee concurrency-safe, gap-free incremental numbers. Clients can easily register new formats and generate IDs at runtime by supplying only the template ID and dynamic data.


## Configuration Schema
```json
{
  "template": "{ORG}-{DATE:yyyyMMdd}-{SEQ}-{RAND}",
  "sequence": {
    "scope": "daily",
    "start": 1,
    "padding": { "length": 4, "char": "0" }
  },
  "random": {
    "length": 2,
    "charset": "A-Z0-9"
  }
}
```
### 1. Template
* Type: string
* Value: A pattern containing static text and dynamic tokens.
* Current tokens:
        {ORG} → a client-supplied organization or tenant code. (Can be any variable)
        {DATE:yyyyMMdd} → the current date in YYYYMMDD format
        {SEQ} → a daily sequence counter
        {RAND} → a random string

Example output (for ORG=PG, on July 3 2025):
```
PG-20250703-0001-A9
```
### 2. Sequence
Controls behavior of the {SEQ} token.

| Property         | Type      | Description                                                                                |
| ---------------- | --------- | ------------------------------------------------------------------------------------------ |
| `scope`          | *string*  | When the counter resets:                                                                   |
|                  |           | • `daily` → resets to `start` at 00:00 local time each day                                 |
| `start`          | *integer* | Initial value when sequence is first used or after each reset (here: `1`)                  |
| `padding.length` | *integer* | Minimum width in characters (here: `4`). Short values are left-padded with `padding.char`. |
| `padding.char`   | *string*  | Character used for padding (here: `"0"`, so `1` → `0001`).                                 |


### 3. Random
Controls behavior of the {RAND} token.

| Property  | Type      | Description                                                             |
| --------- | --------- | ----------------------------------------------------------------------- |
| `length`  | *integer* | Number of characters to generate (here: `2`).                           |
| `charset` | *string*  | Allowed characters. Here `"A-Z0-9"` means uppercase A–Z and digits 0–9. |

### 4. Runtime Evaluation Flow
1. Load configuration by template name.
2. Parse the template into an ordered list of segments (static vs. tokens).
3. For each segment:
    * Static → append verbatim.
    * DATE → format now() with given pattern.
    * SEQ → 
        * Fetch next value from your counter store (DB).
        * Apply padding (4 digits).
        *   If scope="daily" and date changed, reset to start.
    * RAND → pick length chars at random from charset.
4. Concatenate all parts to produce the final ID string.


### Validation & Error Handling

* Malformed date pattern → fail at config-load time.
* Padding shorter than digits in start → validation error on startup.
* Counter store unavailable → either block until recovery .



## Function Signatures
```
registerTemplate(templateId: string, config: object): void
generateId(templateId: string, variables: Map<string, string>): string
```

### Function: Register Template Config

```
registerTemplate(templateId, config)
```
#### Parameters:
| Param        | Type        | Description                                            |
| ------------ | ----------- | ------------------------------------------------------ |
| `templateId` | String      | Unique identifier(code) for this template (e.g. `"pt.assessmentnumber"`) |
| `config`     | JSON object | Template config JSON (as described earlier)            |


#### Responsibilities:
Store the template config in a database table, for example:
```
CREATE TABLE idgen_templates (
  id            character varying(64) PRIMARY KEY,
  config        JSONB NOT NULL,
  created_at    bigint,
  created_by    character varying(64)
);
```
Create a Postgres sequence:
```
CREATE SEQUENCE IF NOT EXISTS seq_{templateId}
  START WITH config.sequence.start
  INCREMENT BY 1
  MINVALUE 1
  CACHE 1;
```

#### Example usage
````
registerTemplate("orderId", {
  "template": "{ORG}-{DATE:yyyyMMdd}-{SEQ<4>}-{RAND<2>}",
  "sequence": { "scope": "daily", "start": 1, "padding": { "length": 4, "char": "0" } },
  "random": { "length": 2, "charset": "A-Z0-9" }
});
````


## Function: Generate ID

```
generateId(templateId, variables)
```
#### Parameters:
| Param        | Type                 | Description                                                              |
| ------------ | -------------------- | ------------------------------------------------------------------------ |
| `templateId` | String               | Template identifier (e.g. `"pt.assessmentnumber"`)                                   |
| `variables`  | Map\<String, String> | Map of runtime-supplied values for placeholders (e.g. `{ "ORG": "PG" }`) |

#### Responsibilities
* Fetch the template config JSON from DB
* Parse template string.
* Replace:
    * {VARIABLE} → values from variables map (e.g. {ORG}, {TENANT}, {DEP})
    * {DATE:...} → formatted current date
    * {SEQ} → get nextval from Postgres sequence seq_{templateId}, pad to length 
    * {RAND} → random string from allowed charset
* Concatenate and return the final ID string.

#### Example Usage
```
generateId("orderId", { "ORG": "PG" });
// → PG-20250703-0001-A9
```
