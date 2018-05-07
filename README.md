# resource-align
Web service to align resources to curriculum standards, drawing on multiple criteria

Works from a dummy repository (tab-delimited file), comprising:

* URL: URL of resource
* Content: Textual content of resource, to be aligned to curriculum via https://github.com/nsip/curriculum-align
* Paradata: JSON object, mapping curriculum items to the frequency with which the resource has been used to teach against that curriculum item in the repository
* Manual-Alignment: Nominated curriculum items that the resource has been aligned to in metadata, whether by the resource author, a local expert, or a third party

