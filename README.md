A tool to create/render/search/archive invoices from command line
============

![Imgur](http://i.imgur.com/khOrNjzb.png?1) 

------
[Downloads](#downloads) | [Installation](#installation) | [Workflow](#workflow)
 
------



Motivations
============
 
I am a techie and I am comfortable with the command line and other developer tools (like git). 
So why not manage the invoices from command line and use git to back them up. Also I wanted to 
explore the golang universe and this sounded like a nice project.

Features 
============

- generate pdf file for invoices with automatic calculation of subtotal/taxes/total
- generate and encrypted descriptor of the invoice data to store/backup/recovery of older invoices
- invoices full text search, searchable by customer, amount and date. 
- customize (some) of the layout of the generated pdf
- i18n templates for localisation/customization of invoices pdf content
- export data from [DailyTimeApp](https://dailytimeapp.com/) (mac only)

Concepts
============

__workspace__ is the folder where the invoice pdfs and encrypted descriptors are generated. 
In the __workspace__ folder there is a special file, ```_master.json```, that is the __master descriptor__
for invoices. The __master descriptor__ is the json file used by govoice to render invoice pdfs. 
  
The __configuration__ of *govoice*  is stored in ```$HOME/.govoice```. Inside the __configuration__ folders 
there is the ```config.toml``` file conaining some system properties and the parameters for **pdfs layout**. 
 Also there is a folder called ```i18n``` where there are the **localization** for the invoice contents.
 

Installation
============

*govoice* is a static compiled golang program, 
it does not require any 3rd party software or library
and can be used on windows/linux/mac

**step 1.** download the binary for your system: [TODO link to binaries]

**step 2.** move the executable somewhere in your $PATH (ex. /usr/local/bin)

**step 3.** run ```govoice config --workspace /path/to/the/workspace``` 

**step 4.** edit ```/path/to/the/workspace/_master.json``` with your personal infos
 
Workflow
============

the common workflow for *govoice* is to **edit the master descriptor with the invoice data**, 
this can be done **by running  the command ```govoice edit```**. 
Once your are done, **then run the command ```govoice render```** that will generate the pdf and an encrypted 
copy of the master descriptor in the workspace folder. That's it.

### Using GIT to archive/backup invoices
At this point you could have the workspace folder under vc, committing only the encrypted descriptors. 
An example, using git, of ```.gitignore``` in the workspace is:

```
**/*.json
**/*.pdf
```

### Searching for invoices
Every time an invoice is rendered it is indexed in a local full text search index. 
To search for an invoice the command ```govoice search ...``` can be used. 

### Restore an invoice
In case an invoice needs to be restored (for example for editing) the command ```govoice restore INVOICENUMBER``` 
is provided. The command will replace the __master descriptor__ content with the content of the restored 
invoice

Security considerations
============

*govoice* uses [CFB](https://en.wikipedia.org/wiki/Block_cipher_mode_of_operation#Cipher_Feedback_.28CFB.29) 
to encrypt the invoices ([this implementation](https://golang.org/src/crypto/cipher/example_test.go)), using 
key of 32 bytes.  

[TODO how can I decrypt the info without govoice?]

Invoice descriptor 
============

The invoice descriptor contains the details of each invoice. Usually the first time you will have 
to insert your data and the invoice settings, after that usually what needs to be updated each time is:

- the invoice number `json:{invoice.number}`
- the invoice date `json:{invoice.date}`
- the invoice due `json:{invoice.due}`
- the customer data `json:{to.*}`
- the billed items `json:{items.*}`

here there is the full format of the invoice descriptor:

```
{
  # this section is for the invoce sender 
  "from": {
    "name": "My Name",  
    "address": "My Address",
    "city": "My City",
    "area_code": "My Post Code",
    "country": "My Country",
    "tax_id": "My Tax ID",
    "vat_number": "My VAT Number",
    "email": "My Email"
  },
  # this section is for the customer being invoiced 
  "to": {
    "name": "Customer Name",       <--- this field is indexed for search     
    "address": "Customer Address",
    "city": "Customer City",
    "area_code": "Customer Post Code",
    "country": "Customer Country",
    "tax_id": "Customre Tax ID",
    "vat_number": "Customer VAT number",
    "email": "Customer Email"
  },
  # here there are the bank account details for the payment
  "payment_details": {
    "account_holder": "My Name",
    "account_bank": "My Bank Name",
    "account_iban": "My IBAN",
    "account_bic": "My BIC/SWIFT"
  },
  # here is the invoice data 
  "invoice": {
    "number": "0000000",           <--- this is used to generate the invoice file name (string)
    "date": "01.01.2017",          <--- date of the invoice, it is indexed for search
    "due": "01.02.2017"            <--- date of payment due
  },
  # general settings for the invoice
  "settings": {
    "items_price": 45,             <--- default price of items 
    "items_quantity_symbol": "",   <--- quantity symbol for items (for example h)
    "vat_rate": 19,                <--- vat rate (as percentage)
    "currency_symbol": "€",        <--- currency symbol to use in pdf
    "lang": "en",                  <--- this is used to select the i18n template (en = .govoice/i18n/en.toml) 
    "date_format": "%d.%m.%y"      <--- this is the format of the date of {invoice.date} if not empty overrides the default (see main configuration)
  },
  # this section is to configure dailytimeapp integration
  "dailytime": {                   
    "enabled": false,              <--- enable/disable integration 
    "date_from": "",               <--- date range from of export (beginning) the format is system dependent 
    "date_to": "",                 <--- date range to of export (end) the format is system dependent
    "projects": null               
  },
  # list of items of the invoices 
  "items": [
    {
      "description": "item 1 description",  <--- description of the activity/product
      "quantity": 10                        <--- how much of this item are billed 
    },
    {
      "description": "web development",     <--- description of the activity/product
      "quantity": 5,                        <--- how much of this item are billed
      "price": 60,                          <--- [OPTIONAL] overrides {settings.items_price} for this item
      "quantity_symbol" : "pieces"          <--- [OPTIONAL] overrides {settings.items_quantity_symbol} for this item
    }
  ],
  # list of notes to append to the invoice
  "notes": [
    "first note",
    "second note"
  ]
}

```

#### DailyTimeApp integration
DailyTimeApp is a nice time tracking application for mac that allows to export data in via scripting,
_govoice_ can export the activities from daily to the invoice. When the invoice will be rendered and 
archived, the dailytime integration is disabled and the items are explicitly listed in the invoice descriptor. 

Here is an example to enable dailytimeapp integration:
```
...
"dailytime": {                   
    "enabled": true,              <--- enable/disable integration
    "date_from": "20/01/2017",    <--- date range from of export (beginning) the format is system dependent 
    "date_to": "31/12/2017",      <--- date range to of export (end) the format is system dependent
    "projects": [
      { 
        "name" : "projectX-dev",                   <--- name of the dailytimeapp activity 
        "item_description" : "website development" <--- item description to use for the activity in the invoice
        "item_price"                               <--- [OPTIONAL] overrides {settings.items_price} for this item
      }
    ]
  },
...
```


i18n templates
============

[todo sample with explanation]

Configuration
============

The configuration file is by default stored in `$HOME/.govoice/config.toml`

````
dateInputFormat = "%d/%m/%y"    <--- default date format (can be overridden in invoice descriptor)
masterTemplate = "_master"      <--- name of the master descriptor
searchResultLimit = 50          <--- NOT USED
workspace = "/tmp/govoice"      <--- workspace location (see govoice config --workspace) 
````

Commands & Usage
============

run `govoice help`to see available commands:
 
```
Usage:
  govoice [command]

Available Commands:
  config      configure govoice
  edit        edit the master descriptor using the system editor
  help        Help about any command
  index       (re)generate the searchable index of invoices
  info        print information about paths (when you forget where they are)
  render      render the master invoice in the workspace
  restore     restore a generated (and ecrypted) invoice descriptor to the master descriptor for editing
  search      query the index to search for invoices

Flags:
  -h, --help   help for govoice
```


 
Downloads
============

*govoice* is available for windows/mac/linux (amd64).


##### stable-v0.1.0: [linux](https://gitlab.com/almost_cc/govoice/builds/artifacts/v0.1.0/download?job=build-linux) / [mac](https://gitlab.com/almost_cc/govoice/builds/artifacts/v0.1.0/download?job=build-mac) / [windows](https://gitlab.com/almost_cc/govoice/builds/artifacts/v0.1.0/download?job=build-windows)

##### latest-master: [linux](https://gitlab.com/almost_cc/govoice/builds/artifacts/master/download?job=build-linux) / [mac](https://gitlab.com/almost_cc/govoice/builds/artifacts/master/download?job=build-mac) / [windows](https://gitlab.com/almost_cc/govoice/builds/artifacts/master/download?job=build-windows)


 
Known issues
============

- the command line output for render doesn't print the currency codes correctly 

CI
============

[![build status](https://gitlab.com/almost_cc/govoice/badges/master/build.svg)](https://gitlab.com/almost_cc/govoice/commits/master)


[![coverage report](https://gitlab.com/almost_cc/govoice/badges/master/coverage.svg)](https://gitlab.com/almost_cc/govoice/commits/master)


for code coverage badge
``` coverage:\s*\d+.\d+\%\s+of\s+statements ```