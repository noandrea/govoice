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

Invoice descriptor 
============

[todo sample with explanation]

i18n templates
============

[todo sample with explanation]

Configuration
============

[todo sample with explanation]
 
Downloads
============

*govoice* is available for:
- [linux](https://gitlab.com/almost_cc/govoice/builds/artifacts/master/download?job=build-linux)
- [mac](https://gitlab.com/almost_cc/govoice/builds/artifacts/master/download?job=build-mac)
- [windows](https://gitlab.com/almost_cc/govoice/builds/artifacts/master/download?job=build-windows)

all builds are for amd64
 
Known issues
============

- the command line output for render doesn't print the currency codes correctly 

CI
============

[![build status](https://gitlab.com/almost_cc/govoice/badges/master/build.svg)](https://gitlab.com/almost_cc/govoice/commits/master)


[![coverage report](https://gitlab.com/almost_cc/govoice/badges/master/coverage.svg)](https://gitlab.com/almost_cc/govoice/commits/master)


for code coverage badge
``` coverage:\s*\d+.\d+\%\s+of\s+statements ```