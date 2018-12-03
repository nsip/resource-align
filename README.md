# resource-align
Web service to align resources to curriculum standards, drawing on multiple criteria

NOTE: This is experimental and proof-of-concept code

This code builds on https://github.com/nsip/curriculum-align (which it also invokes), and it proposes resources from a repository that are aligned to a given set of learning areas, year levels, and (optionally) curriculum standards, in ranked order.

Binary distributions of the code are available in the build/ directory.

The web service is made available as a library (`Align()`); the `cmd` directory contains a sample shell for it, which is used in the binary distribution. In the sample shell, the web service runs on port 1576. The test script `test.sh` issues representative REST queries against the web service.

The web service takes the following arguments:

````
GET http://localhost:1576/align?yr=X1,X2,X3&area=Y1,Y2,Y3&item=Z1,Z2,Z3....
````

where yr is the year level, area is the learning area, and item is the curriculum standard to be aligned. All of these can have multiple comma-delimited values. All parameters are optional; if no parameters are given, alignment is attempted against all curriculum items configured in the https://github.com/nsip/curriculum-align instance.

The repository that this implementation works from is a set of JSON files in the `repository` folder of the executable; in this distribution, sample files are in `cmd/repository`. The JSON structure required is an array of JSON objects, with the following structure:

* URL: URL of resource
* Content: Textual content of resource, to be used for alignment to curriculum via https://github.com/nsip/curriculum-align. (NOTE: the value of Content needs to be kept to a single line.)
* Paradata: JSON object, mapping curriculum items codes to the frequency with which the resource has been used to teach against that curriculum item in the repository
* Manual-Alignment: List of nominated curriculum item codes that the resource has been aligned to in metadata, whether by the resource author, a local expert, or a third party
* Learning-Area: List of learning areas that the resource relates to; is expected to be the same set of values as used in the related curriculum
* Year: List of year levels that the resource relates to; is expected to be the same set of values as used in the related curriculum

For example:

````
  {
    "URL":"http://www.skwirk.com/p-u_s-4_u-198/planet-earth/nsw/science",
    "Content":"Chapter 1: The layers of the Earth Chapter 2: The tectonic plates Chapter 3: Volcanoes: birth, life and death Chapter 4: Earthquakes: cause, reaction and measuring Topic 2 : Rocks and minerals Chapter 1: Minerals, crystals and ores Chapter 2: Igneous rocks Chapter 3: Sedimentary rocks Chapter 4: Metamorphic rocks Topic 3 : The atmosphere Chapter 1: The layers of air Chapter 2: Weather and climate Chapter 3: The greenhouse effect Chapter 4: The ozone layer Topic 4 : The hydrosphere Chapter 1: The water cycle Chapter 2: Tides Topic 5 : Shaping the Earth Chapter 1: Weathering: chemical and physical Chapter 2: Erosion: people and nature",
    "Paradata":{
      "ACSSU153":4,
      "ACSIS148":1,
      "SC4_12ES":2,
      "SC4-3VA":1
    },
    "Manual-Alignment":[
      "ACSSU153",
      "SC4_12ES"
    ],
    "Learning-Area": ["Science"],
    "Year": ["7", "8"]
  },
````

The response is a JSON list of structs, one for each resource matched, with the following fields:

* URL: The URL of the resource
* Content: The abstract of the resource
* Statement: The identifier of the curriculum item for which this resource is the best match. The curriculum items matched against are filtered by the parameters given in the web service call.
* Expert: The score for the best match based on expert advice (the Manual-Alignment field of the repository entry).
* Usage: The score for the best match based on usage (the Paradata field of the repository entry).
* TextBased: The score for the best match based on keyword alignment (calling curriculum-align with the Content field of the repository entry).
* WeightedTotal: the overall score for the best match; currently the sum of the Expert, Usage, and TextBased scores.

Each of the Expert, Usage, and TextBased scores are normalised: for each item, they are normalised to range from 0 to 1. The response is ranked in order of WeightedTotal.

As with curriculum-align, all text tokenised in the gem from the resource repository is also indexed and available for retrieval:

````
GET http://localhost:1576/index?search=word
````

