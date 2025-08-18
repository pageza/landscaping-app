const puppeteer = require('puppeteer');

class LinkTester {
    constructor() {
        this.browser = null;
        this.page = null;
        this.baseUrl = 'http://localhost:9090';
        this.visited = new Set();
        this.results = {
            working: [],
            broken: [],
            redirects: [],
            summary: {}
        };
    }

    async init() {
        this.browser = await puppeteer.launch({ 
            headless: false,
            defaultViewport: { width: 1280, height: 720 }
        });
        this.page = await this.browser.newPage();
        
        // Listen for console logs and errors
        this.page.on('console', msg => {
            if (msg.type() === 'error') {
                console.log('❌ Page Error:', msg.text());
            }
        });
        
        this.page.on('pageerror', error => {
            console.log('❌ JavaScript Error:', error.message);
        });
    }

    async getAllLinks(currentUrl) {
        try {
            console.log(`\n🔍 Scanning page: ${currentUrl}`);
            await this.page.goto(currentUrl, { waitUntil: 'networkidle0', timeout: 10000 });
            
            // Wait a bit for any dynamic content to load
            await this.page.waitForTimeout(2000);
            
            // Get all links on the page
            const links = await this.page.evaluate(() => {
                const anchors = Array.from(document.querySelectorAll('a[href]'));
                return anchors.map(anchor => ({
                    href: anchor.href,
                    text: anchor.textContent.trim(),
                    innerHTML: anchor.innerHTML.trim(),
                    selector: anchor.tagName + (anchor.id ? `#${anchor.id}` : '') + (anchor.className ? `.${anchor.className.split(' ').join('.')}` : ''),
                    position: {
                        x: anchor.offsetLeft,
                        y: anchor.offsetTop
                    }
                }));
            });

            console.log(`   Found ${links.length} links on this page`);
            return links;
        } catch (error) {
            console.log(`❌ Error scanning page ${currentUrl}:`, error.message);
            return [];
        }
    }

    async testLink(link, sourcePage) {
        try {
            console.log(`   Testing: "${link.text}" → ${link.href}`);
            
            // Skip external links, javascript, mailto, tel links
            if (link.href.startsWith('javascript:') || 
                link.href.startsWith('mailto:') || 
                link.href.startsWith('tel:') ||
                link.href.startsWith('http') && !link.href.includes('localhost')) {
                
                this.results.working.push({
                    url: link.href,
                    text: link.text,
                    sourcePage: sourcePage,
                    status: 'external/special',
                    note: 'External link or special protocol'
                });
                return;
            }

            // Try to navigate to the link
            const response = await this.page.goto(link.href, { 
                waitUntil: 'networkidle0', 
                timeout: 10000 
            });
            
            await this.page.waitForTimeout(1000);

            const finalUrl = this.page.url();
            const statusCode = response ? response.status() : 'unknown';
            
            // Check if page loaded successfully
            const pageTitle = await this.page.title();
            const bodyText = await this.page.evaluate(() => document.body ? document.body.innerText.substring(0, 200) : 'No body content');
            
            if (statusCode >= 200 && statusCode < 400) {
                // Check for error indicators in the page content
                const hasError = bodyText.toLowerCase().includes('error') || 
                               bodyText.toLowerCase().includes('not found') ||
                               bodyText.toLowerCase().includes('404') ||
                               pageTitle.toLowerCase().includes('error');
                
                if (hasError) {
                    this.results.broken.push({
                        url: link.href,
                        finalUrl: finalUrl,
                        text: link.text,
                        sourcePage: sourcePage,
                        status: statusCode,
                        pageTitle: pageTitle,
                        bodyPreview: bodyText,
                        issue: 'Page loaded but contains error content'
                    });
                } else if (finalUrl !== link.href) {
                    this.results.redirects.push({
                        url: link.href,
                        finalUrl: finalUrl,
                        text: link.text,
                        sourcePage: sourcePage,
                        status: statusCode,
                        pageTitle: pageTitle
                    });
                } else {
                    this.results.working.push({
                        url: link.href,
                        text: link.text,
                        sourcePage: sourcePage,
                        status: statusCode,
                        pageTitle: pageTitle
                    });
                }
            } else {
                this.results.broken.push({
                    url: link.href,
                    finalUrl: finalUrl,
                    text: link.text,
                    sourcePage: sourcePage,
                    status: statusCode,
                    pageTitle: pageTitle,
                    bodyPreview: bodyText,
                    issue: `HTTP ${statusCode} error`
                });
            }

        } catch (error) {
            this.results.broken.push({
                url: link.href,
                text: link.text,
                sourcePage: sourcePage,
                status: 'error',
                issue: error.message,
                errorType: error.name
            });
        }
    }

    async testAllLinksOnPage(pageUrl) {
        if (this.visited.has(pageUrl)) {
            console.log(`⏭️  Skipping already visited: ${pageUrl}`);
            return [];
        }
        
        this.visited.add(pageUrl);
        const links = await this.getAllLinks(pageUrl);
        
        for (const link of links) {
            await this.testLink(link, pageUrl);
        }
        
        // Return internal links for further crawling
        return links.filter(link => 
            link.href.includes('localhost') && 
            !link.href.startsWith('javascript:') &&
            !link.href.startsWith('mailto:') &&
            !link.href.startsWith('tel:')
        );
    }

    async runFullSiteTest() {
        console.log('🚀 Starting comprehensive link testing...\n');
        
        // Start with main pages
        const startPages = [
            `${this.baseUrl}/`,
            `${this.baseUrl}/services`,
            `${this.baseUrl}/about`,
            `${this.baseUrl}/contact`,
            `${this.baseUrl}/login`,
            `${this.baseUrl}/signup`,
            `${this.baseUrl}/booking`
        ];

        const pagesToVisit = [...startPages];
        const discoveredPages = new Set();

        while (pagesToVisit.length > 0) {
            const currentPage = pagesToVisit.shift();
            
            if (this.visited.has(currentPage)) continue;
            
            console.log(`\n📄 Testing page: ${currentPage}`);
            const internalLinks = await this.testAllLinksOnPage(currentPage);
            
            // Add new internal pages to visit
            for (const link of internalLinks) {
                if (!this.visited.has(link.href) && !discoveredPages.has(link.href)) {
                    discoveredPages.add(link.href);
                    pagesToVisit.push(link.href);
                }
            }
        }

        // Test admin pages if we can authenticate
        await this.testAdminPages();
    }

    async testAdminPages() {
        try {
            console.log('\n🔐 Testing admin pages...');
            
            // Go to login page
            await this.page.goto(`${this.baseUrl}/login`);
            await this.page.waitForTimeout(1000);
            
            // Try to login as admin
            await this.page.type('input[name="email"]', 'admin@landscapepro.com');
            await this.page.type('input[name="password"]', 'password123');
            await this.page.click('button[type="submit"]');
            await this.page.waitForTimeout(2000);
            
            // Check if login was successful
            const currentUrl = this.page.url();
            if (currentUrl.includes('login')) {
                console.log('❌ Could not login to admin, skipping admin pages');
                return;
            }
            
            // Test admin pages
            const adminPages = [
                `${this.baseUrl}/admin`,
                `${this.baseUrl}/admin/customers`,
                `${this.baseUrl}/admin/jobs`,
                `${this.baseUrl}/admin/services`,
                `${this.baseUrl}/admin/team`,
                `${this.baseUrl}/admin/reports`
            ];
            
            for (const adminPage of adminPages) {
                await this.testAllLinksOnPage(adminPage);
            }
            
        } catch (error) {
            console.log('❌ Error testing admin pages:', error.message);
        }
    }

    generateReport() {
        this.results.summary = {
            totalLinks: this.results.working.length + this.results.broken.length + this.results.redirects.length,
            workingLinks: this.results.working.length,
            brokenLinks: this.results.broken.length,
            redirects: this.results.redirects.length,
            pagesVisited: this.visited.size
        };

        console.log('\n' + '='.repeat(80));
        console.log('📊 LINK TESTING REPORT');
        console.log('='.repeat(80));
        
        console.log(`\n📈 Summary:`);
        console.log(`   Pages visited: ${this.results.summary.pagesVisited}`);
        console.log(`   Total links tested: ${this.results.summary.totalLinks}`);
        console.log(`   ✅ Working links: ${this.results.summary.workingLinks}`);
        console.log(`   🔄 Redirects: ${this.results.summary.redirects}`);
        console.log(`   ❌ Broken links: ${this.results.summary.brokenLinks}`);

        if (this.results.broken.length > 0) {
            console.log(`\n❌ BROKEN LINKS (${this.results.broken.length}):`);
            console.log('-'.repeat(80));
            this.results.broken.forEach((link, index) => {
                console.log(`${index + 1}. Link: "${link.text}"`);
                console.log(`   URL: ${link.url}`);
                console.log(`   Source Page: ${link.sourcePage}`);
                console.log(`   Issue: ${link.issue || link.status}`);
                console.log(`   Page Title: ${link.pageTitle || 'N/A'}`);
                if (link.bodyPreview) {
                    console.log(`   Page Content: ${link.bodyPreview.substring(0, 100)}...`);
                }
                console.log('');
            });
        }

        if (this.results.redirects.length > 0) {
            console.log(`\n🔄 REDIRECTS (${this.results.redirects.length}):`);
            console.log('-'.repeat(80));
            this.results.redirects.forEach((link, index) => {
                console.log(`${index + 1}. Link: "${link.text}"`);
                console.log(`   Original: ${link.url}`);
                console.log(`   Redirected to: ${link.finalUrl}`);
                console.log(`   Source Page: ${link.sourcePage}`);
                console.log('');
            });
        }

        // Save detailed report to file
        const reportData = {
            timestamp: new Date().toISOString(),
            summary: this.results.summary,
            brokenLinks: this.results.broken,
            redirects: this.results.redirects,
            allWorkingLinks: this.results.working
        };

        require('fs').writeFileSync('link-test-report.json', JSON.stringify(reportData, null, 2));
        console.log('\n💾 Detailed report saved to: link-test-report.json');
    }

    async close() {
        if (this.browser) {
            await this.browser.close();
        }
    }

    async run() {
        try {
            await this.init();
            await this.runFullSiteTest();
            this.generateReport();
        } catch (error) {
            console.error('❌ Test runner error:', error);
        } finally {
            await this.close();
        }
    }
}

// Run the test
const tester = new LinkTester();
tester.run().then(() => {
    console.log('\n✅ Link testing completed!');
    process.exit(0);
}).catch(error => {
    console.error('❌ Link testing failed:', error);
    process.exit(1);
});