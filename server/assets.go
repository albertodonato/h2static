package server

// AssetDetails holds details for a static asset.
type AssetDetails struct {
	ContentType string
	Content     []byte
}

// StaticAssets maps static assets to their details.
type StaticAssets map[string]AssetDetails

// Static assets for the application.
var staticAssets = StaticAssets{
	"logo.svg": {
		ContentType: "image/svg+xml",
		Content: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<svg width="100px" height="100px" version="1.1" viewBox="0 0 23.083 28.86" xmlns="http://www.w3.org/2000/svg">
 <g transform="translate(-59.39 -143.13)" stroke="#fff">
  <g fill="#90bae3" stroke-width=".2" style="font-feature-settings:normal;font-variant-caps:normal;font-variant-ligatures:normal;font-variant-numeric:normal" aria-label="//">
   <path d="m69.192 147.62q0.33752-1.0126 1.3808-1.0126h0.03068q0.72108 0 1.166 0.56766 0.42958 0.61368 0.19945 1.3194l-6.1828 19.009q-0.30684 1.0126-1.3808 1.0126h-0.04603q-0.73642 0-1.1507-0.56766-0.41424-0.55231-0.19945-1.2887z" style="font-feature-settings:normal;font-variant-caps:normal;font-variant-ligatures:normal;font-variant-numeric:normal"/>
   <path d="m77.538 147.62q0.33752-1.0126 1.3808-1.0126h0.03068q0.72108 0 1.166 0.56766 0.42958 0.61368 0.19945 1.3194l-6.1828 19.009q-0.30684 1.0126-1.3808 1.0126h-0.04603q-0.73642 0-1.1507-0.56766-0.41424-0.55231-0.19945-1.2887z" style="font-feature-settings:normal;font-variant-caps:normal;font-variant-ligatures:normal;font-variant-numeric:normal"/>
  </g>
  <g transform="skewX(-17.8)" fill="#449c1e" stroke-linecap="square" stroke-linejoin="round" style="font-feature-settings:normal;font-variant-caps:normal;font-variant-ligatures:normal;font-variant-numeric:normal" aria-label="h:s">
   <path d="m117.75 156.8v5.2865h-2.7905v-4.0308q0-1.1395-0.0543-1.5658-0.0465-0.42633-0.17053-0.62787-0.16278-0.2713-0.44184-0.41858-0.27905-0.15503-0.63562-0.15503-0.86816 0-1.3642 0.67438-0.4961 0.66663-0.4961 1.8526v4.2711h-2.775v-12.061h2.775v4.6509q0.62787-0.75965 1.3332-1.1162 0.70539-0.36432 1.558-0.36432 1.5038 0 2.2789 0.92242 0.7829 0.92243 0.7829 2.682z" stroke-width=".2" style="font-feature-settings:normal;font-variant-caps:normal;font-variant-ligatures:normal;font-variant-numeric:normal;paint-order:markers fill stroke"/>
   <path d="m120.78 153.4h2.7983v2.9998h-2.7983zm0 5.6818h2.7983v2.9998h-2.7983z" stroke-width=".2" style="font-feature-settings:normal;font-variant-caps:normal;font-variant-ligatures:normal;font-variant-numeric:normal;paint-order:markers fill stroke"/>
   <path d="m133.48 153.67v2.1084q-0.89142-0.37207-1.7208-0.55811-0.8294-0.18603-1.5658-0.18603-0.79065 0-1.1782 0.20153-0.37983 0.19379-0.37983 0.60462 0 0.33331 0.28681 0.51159 0.29455 0.17829 1.0464 0.26355l0.48834 0.0698q2.1316 0.2713 2.868 0.89142 0.73639 0.62011 0.73639 1.9456 0 1.3875-1.0232 2.0852-1.0232 0.69763-3.0541 0.69763-0.86041 0-1.7828-0.13953-0.91467-0.13177-1.8836-0.40307v-2.1084q0.8294 0.40307 1.6976 0.60461 0.87591 0.20154 1.7751 0.20154 0.81391 0 1.2247-0.22479 0.41083-0.22479 0.41083-0.66663 0-0.37207-0.2868-0.55035-0.27906-0.18604-1.124-0.28681l-0.48834-0.062q-1.8526-0.23254-2.5967-0.86041t-0.74414-1.9069q0-1.3798 0.94568-2.0464 0.94568-0.66662 2.899-0.66662 0.76739 0 1.6123 0.11627t1.8371 0.36432z" stroke-width=".2" style="font-feature-settings:normal;font-variant-caps:normal;font-variant-ligatures:normal;font-variant-numeric:normal;paint-order:markers fill stroke"/>
  </g>
 </g>
</svg>
`),
	},

	"style.css": {
		ContentType: "text/css",
		Content: []byte(`
body {
  width: 90%;
  margin: 0 auto;
  font-family: sans;
  font-size: 34px;
  color: black;
}
h1 {
  margin: 1em 0;
  font-size: 130%;
}
a, a:visited {
  color: black;
  text-decoration: none;
}
a:active, a:hover {
  color: #007bff;
  text-decoration: none;
}
.logo {
  display: inline-block;
  vertical-align: middle;
}
.logo img {
  width: 3em;
  height: 3em;
}
.title {
 margin-left: 0.5em;
}
.listing {
  width: 100%;
}
.row {
  padding: 0.5rem 0;
  display: flex;
  justify-content: space-between;
}
.col {
  display: inline-block;
  margin: 0 0.2rem;
  padding: 1rem 0.5rem;
  font-family: monospace;
  border-width: 1px;
  border-style: solid;
  border-radius: 0.25rem;
  white-space: nowrap;
}
a.type-dir-up {
  flex-grow: 0;
  width: auto;
  background: #6c757d linear-gradient(to bottom, #828a91 0, #6c757d 100%);
  border-color: #6c757d;
  color: white;
}
a.type-dir {
  background: #337ab7 linear-gradient(to bottom, #337ab7 0, #2e6da4 100%);
  border-color: #337ab7;
  color: white;
}
a.type-file {
  background: #dddddd linear-gradient(to bottom, #f5f5f5 0, #e8e8e8 100%);
  border-color: #dddddd;
  color: #515151;
}
.sort a {
  background: #6c757d linear-gradient(to bottom, #828a91 0, #6c757d 100%);
  border-color: #6c757d;
  color: white;
  font-size: 80%;
}
.sort-asc .col-name.sorted::after,
.sort-asc .col-size.sorted::before {
  margin: 0 0.5em;
  content: "\0025B2";
}
.sort-desc .col-name.sorted::after,
.sort-desc .col-size.sorted::before {
  margin: 0 0.5em;
  content: "\0025BC";
}
.path {
  font-family: monospace;
}
.col-name {
  flex-grow: 1;
}
.col-size {
  border-color: #777777;
  background-color: white;
  color: #777777;
  text-align: right;
  width: 11rem;
}
.size-suffix {
  display: inline-block;
  width: 1.5em;
  margin-left: 0.25em;
  font-size: 80%;
  text-align: left;
}
.powered-by {
  margin: 3em 0;
  text-align: center;
  font-size: 80%;
}
.powered-by a {
  font-family: monospace;
  font-size: 120%;
  margin-left: 0.5em;
}
a.powered-by:hover {
  text-decoration: underline;
}

@media (min-width: 992px) {
  body {
    width: 60%;
    font-size: 16px;
  }
  .logo img {
    width: 4em;
    height: 4em;
  }
  .col {
    padding: 0.5rem;
  }
  .col-size {
    width: 5rem;
  }
}
`),
	},
}
