package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `
<!DOCTYPE html>
<html>
<head>
    <title>LandscapePro</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="max-w-4xl mx-auto py-12">
        <h1 class="text-4xl font-bold text-green-600 text-center mb-8">🌿 LandscapePro</h1>
        <div class="bg-white p-8 rounded-lg shadow-lg">
            <p class="text-xl text-gray-700 mb-6">Professional landscaping services for your home and business.</p>
            
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-8">
                <div class="p-4 bg-green-50 rounded-lg">
                    <h3 class="font-bold">🌱 Lawn Care</h3>
                    <p>Professional mowing and maintenance</p>
                    <p class="text-green-600 font-semibold">From $50/visit</p>
                </div>
                <div class="p-4 bg-blue-50 rounded-lg">
                    <h3 class="font-bold">🌺 Garden Design</h3>
                    <p>Custom landscape design and installation</p>
                    <p class="text-blue-600 font-semibold">Free consultation</p>
                </div>
                <div class="p-4 bg-yellow-50 rounded-lg">
                    <h3 class="font-bold">🌳 Tree Service</h3>
                    <p>Trimming, removal, and health assessment</p>
                    <p class="text-yellow-600 font-semibold">From $200</p>
                </div>
                <div class="p-4 bg-purple-50 rounded-lg">
                    <h3 class="font-bold">💧 Irrigation</h3>
                    <p>Sprinkler system design and repair</p>
                    <p class="text-purple-600 font-semibold">From $150</p>
                </div>
            </div>
            
            <div class="text-center">
                <button 
                    hx-get="/quote" 
                    hx-target="#quote-result"
                    class="bg-green-600 text-white px-8 py-3 rounded-lg hover:bg-green-700">
                    Get Quick Quote
                </button>
                <div id="quote-result" class="mt-4"></div>
            </div>
        </div>
    </div>
</body>
</html>`
		
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, html)
	})

	http.HandleFunc("/quote", func(w http.ResponseWriter, r *http.Request) {
		html := `
		<div class="bg-green-100 p-4 rounded-lg">
			<h3 class="text-xl font-bold text-green-800 mb-2">Estimated Quote: $250 - $350</h3>
			<p class="text-green-700">Based on average property size. Contact us for accurate pricing!</p>
			<p class="text-sm text-green-600 mt-2">📞 (555) 123-4567 | ✉️ info@landscapepro.com</p>
		</div>`
		
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, html)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("🌿 Simple LandscapePro site running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}