package coda

import (
	"fmt"
	"time"
)

/*
	{
	  "id": "AbCDeFGH",
	  "type": "doc",
	  "href": "https://coda.io/apis/v1/docs/AbCDeFGH",
	  "browserLink": "https://coda.io/d/_dAbCDeFGH",
	  "icon": {
		"name": "string",
		"type": "string",
		"browserLink": "https://cdn.coda.io/icons/png/color/icon-32.png"
	  },
	  "name": "Product Launch Hub",
	  "owner": "user@example.com",
	  "ownerName": "Some User",
	  "docSize": {
		"totalRowCount": 31337,
		"tableAndViewCount": 42,
		"pageCount": 10,
		"overApiSizeLimit": false
	  },
	  "sourceDoc": {
		"id": "AbCDeFGH",
		"type": "doc",
		"href": "https://coda.io/apis/v1/docs/AbCDeFGH",
		"browserLink": "https://coda.io/d/_dAbCDeFGH"
	  },
	  "createdAt": "2018-04-11T00:18:57.946Z",
	  "updatedAt": "2018-04-11T00:18:57.946Z",
	  "published": {
		"description": "Hello World!",
		"browserLink": "https://coda.io/@coda/hello-world",
		"imageLink": "string",
		"discoverable": true,
		"earnCredit": true,
		"mode": "view",
		"categories": [
		  "Project Management"
		]
	  },
	  "folder": {
		"id": "fl-1Ab234",
		"type": "folder",
		"browserLink": "https://coda.io/docs?folderId=fl-1Ab234",
		"name": "My docs"
	  },
	  "workspace": {
		"id": "ws-1Ab234",
		"type": "workspace",
		"organizationId": "org-2Bc456",
		"browserLink": "https://coda.io/docs?workspaceId=ws-1Ab234",
		"name": "My workspace"
	  },
	  "workspaceId": "ws-1Ab234",
	  "folderId": "fl-1Ab234"
	}
*/
type Doc struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	WorkspaceID string `json:"workspaceId"`
	FolderID    string `json:"folderId"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (d *Doc) String() string {
	return d.ID + " " + d.Name
}

/*
	{
	    {
	      "id": "canvas-IjkLmnO",
	      "type": "page",
	      "href": "https://coda.io/apis/v1/docs/AbCDeFGH/pages/canvas-IjkLmnO",
	      "browserLink": "https://coda.io/d/_dAbCDeFGH/Launch-Status_sumnO",
	      "name": "Launch Status",
	      "subtitle": "See the status of launch-related tasks.",
	      "icon": {
	        "name": "string",
	        "type": "string",
	        "browserLink": "https://cdn.coda.io/icons/png/color/icon-32.png"
	      },
	      "image": {
	        "browserLink": "https://codahosted.io/docs/nUYhlXysYO/blobs/bl-lYkYKNzkuT/3f879b9ecfa27448",
	        "type": "string",
	        "width": 800,
	        "height": 600
	      },
	      "parent": {
	        "id": "canvas-IjkLmnO",
	        "type": "page",
	        "href": "https://coda.io/apis/v1/docs/AbCDeFGH/pages/canvas-IjkLmnO",
	        "browserLink": "https://coda.io/d/_dAbCDeFGH/Launch-Status_sumnO",
	        "name": "Launch Status"
	      },
	      "children": [
	        {
	          "id": "canvas-IjkLmnO",
	          "type": "page",
	          "href": "https://coda.io/apis/v1/docs/AbCDeFGH/pages/canvas-IjkLmnO",
	          "browserLink": "https://coda.io/d/_dAbCDeFGH/Launch-Status_sumnO",
	          "name": "Launch Status"
	        }
	      ],
	      "authors": [
	        {
	          "@context": "http://schema.org/",
	          "@type": "ImageObject",
	          "additionalType": "string",
	          "name": "Alice Atkins",
	          "email": "alice@atkins.com"
	        }
	      ]
	    }
	}
*/
type Page struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Subtitle string `json:"subtitle"`
	Parent   *Page
	Children []*Page
}

func (p *Page) String() string {
	return p.ID + " " + p.Name
}

/*
	{
	  "id": "grid-pqRst-U",
	  "type": "table",
	  "tableType": "table",
	  "href": "https://coda.io/apis/v1/docs/AbCDeFGH/tables/grid-pqRst-U",
	  "browserLink": "https://coda.io/d/_dAbCDeFGH/#Teams-and-Tasks_tpqRst-U",
	  "name": "Tasks",
	  "parent": {
	    "id": "canvas-IjkLmnO",
	    "type": "page",
	    "href": "https://coda.io/apis/v1/docs/AbCDeFGH/pages/canvas-IjkLmnO",
	    "browserLink": "https://coda.io/d/_dAbCDeFGH/Launch-Status_sumnO",
	    "name": "Launch Status"
	  }
	}
*/
type Table struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	TableType string `json:"tableType"`
	Parent    *Page  `json:"parent"`
}

func (t *Table) String() string {
	return t.ID + " " + t.Name
}

/*
	{
	  "id": "i-tuVwxYz",
	  "type": "row",
	  "href": "https://coda.io/apis/v1/docs/AbCDeFGH/tables/grid-pqRst-U/rows/i-RstUv-W",
	  "name": "Apple",
	  "index": 7,
	  "browserLink": "https://coda.io/d/_dAbCDeFGH#Teams-and-Tasks_tpqRst-U/_rui-tuVwxYz",
	  "createdAt": "2018-04-11T00:18:57.946Z",
	  "updatedAt": "2018-04-11T00:18:57.946Z",
	  "values": {
	    "c-tuVwxYz": "Apple",
	    "c-bCdeFgh": [
	      "$12.34",
	      "$56.78"
	    ]
	  }
	}
*/
type Row struct {
	ID     string         `json:"row"`
	Name   string         `json:"name"`
	Index  int            `json:"index"`
	Values map[string]any `json:"values"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (r *Row) String() string {
	s := r.ID + " | " + r.Name
	if len(r.Values) > 0 {
		s += " |"
		for _, val := range r.Values {
			s += " " + fmt.Sprint(val) + " |"
		}
	}
	return s
}

/*
	{
	  "id": "c-tuVwxYz",
	  "type": "column",
	  "href": "https://coda.io/apis/v1/docs/AbCDeFGH/tables/grid-pqRst-U/columns/c-tuVwxYz",
	  "name": "Completed",
	  "display": true,
	  "calculated": true,
	  "formula": "thisRow.Created()",
	  "defaultValue": "Test",
	  "format": {
	    "type": "text",
	    "isArray": true,
	    "label": "Click me",
	    "disableIf": "False()",
	    "action": "OpenUrl(\"www.google.com\")"
	  }
	}
*/
type Column struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	IsDisplay    bool   `json:"display"`
	IsCalculated bool   `json:"calculated"`
	DefaultValue string `json:"defaultValue"`

	Format struct {
		Type    string `json:"type"`
		IsArray bool   `json:"isArray"`
		Label   string `json:"label"`
	} `json:"format"`
}

func (c *Column) String() string {
	s := c.ID + " " + c.Name + " " + c.Format.Label + " " + c.Format.Type
	if c.Format.IsArray {
		s += " (array)"
	}
	if c.IsDisplay {
		s += " (display)"
	}
	if c.IsCalculated {
		s += " (calculated)"
	}
	return s
}
