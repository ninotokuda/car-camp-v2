<template>
  <div class="about">
    <h1>Map</h1>
    <div id="mapContainer" class="basemap"></div>
  </div>
</template>

<script>
import mapboxgl from "mapbox-gl";
export default {
  name: "Map",
  data() {
    return {
      accessToken:
        "pk.eyJ1Ijoibmlub3Rva3VkYSIsImEiOiJjazl5N3g1NjUwaTJqM3FxZGoxbmh6ZXdtIn0.wGOEHFDgRW3ObcEyCMExyQ",
      map: null
    };
  },
  mounted() {
    mapboxgl.accessToken = this.accessToken;
    console.log("-----");
    this.map = new mapboxgl.Map({
      container: "mapContainer",
      style: "mapbox://styles/ninotokuda/ckheu29os10kq19o61qma4m8h",
      center: [139.52696800231934, 35.355876192506344],
      zoom: 12
    });

    this.map.on("click", e => {
      console.log("click", this.map, e);
      const features = this.map.queryRenderedFeatures(e.point, {layers: ["car-camp-v1"]});
      console.log("----", features);

      if(features.length > 0) {

        const ft = features[0];
        this.showPopup(ft);

      }

		});
		
  },
  methods: {

    showPopup(ft) {

      const spotName = ft.properties.Name;

      const cords = ft.geometry.coordinates.slice();
      const popup = new mapboxgl.Popup({anchor: 'bottom'})
        .setLngLat(cords)
        .setHTML(`<h1>${spotName}</h1>`)
        .addTo(this.map);

        console.log("pop", popup);
    }

    
  }
}
</script>

<style scoped>
.basemap{
	height: 80vh;
}
.popup {
  background-color: red;
}
</style>
