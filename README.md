# JMdict to database converter

A project where I aim to convert the JMdict dictionary file from its glorious
XML format into a database, and hopefully learn a thing or two about Go,
databases, and ORMs in the process.

JMdict is a Japanese/English dictionary project begun by Jim Breen in 1991.
More information and the dictionary file can be found
[here](https://www.edrdg.org/wiki/index.php/JMdict-EDICT_Dictionary_Project).

A tiny subset of the dictionary can be found in `sample.xml`, which contains the
XML DTD and a few entries.

## Notes

- This program has been written for the English-only JMdict file. It is missing
  some fields that are only found in the full dictionary file.
- Order is important to some dictionary elements. However the output from this
  program does not record this, and makes no guarantee that order is preserved
