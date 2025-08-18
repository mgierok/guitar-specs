// Import HTMX and Alpine.js
import 'htmx.org'
import Alpine from 'alpinejs'

// Initialize Alpine.js
window.Alpine = Alpine
Alpine.start()

// HTMX event handlers
document.addEventListener('htmx:beforeRequest', (event) => {
  // Show loading state
  const target = event.target
  if (target) {
    target.classList.add('opacity-50', 'pointer-events-none')
  }
})

document.addEventListener('htmx:afterRequest', (event) => {
  // Hide loading state
  const target = event.target
  if (target) {
    target.classList.remove('opacity-50', 'pointer-events-none')
  }
})

document.addEventListener('htmx:responseError', (event) => {
  // Handle errors
  console.error('HTMX request failed:', event.detail)
})

// Custom utility functions
window.guitarSpecs = {
  // Format guitar features for display
  formatFeatureValue(feature) {
    if (feature.valueText) return feature.valueText
    if (feature.valueNumber) return feature.valueNumber + (feature.unit ? ` ${feature.unit}` : '')
    if (feature.valueBoolean !== null) return feature.valueBoolean ? 'Yes' : 'No'
    if (feature.enumLabel) return feature.enumLabel
    return feature.valueDisplay || 'N/A'
  },
  
  // Toggle feature details with smooth animation
  toggleFeatureDetails(featureId) {
    const details = document.getElementById(`feature-${featureId}`)
    if (details) {
      const isHidden = details.classList.contains('hidden')
      
      if (isHidden) {
        // Show details with smooth animation
        details.classList.remove('hidden')
        details.style.maxHeight = '0'
        details.style.overflow = 'hidden'
        details.style.transition = 'max-height 0.3s ease-out'
        
        // Trigger reflow
        details.offsetHeight
        
        details.style.maxHeight = details.scrollHeight + 'px'
        
        // Clean up after animation
        setTimeout(() => {
          details.style.maxHeight = ''
          details.style.overflow = ''
          details.style.transition = ''
        }, 300)
      } else {
        // Hide details with smooth animation
        details.style.maxHeight = details.scrollHeight + 'px'
        details.style.overflow = 'hidden'
        details.style.transition = 'max-height 0.3s ease-in'
        
        // Trigger reflow
        details.offsetHeight
        
        details.style.maxHeight = '0'
        
        // Hide after animation
        setTimeout(() => {
          details.classList.add('hidden')
          details.style.maxHeight = ''
          details.style.overflow = ''
          details.style.transition = ''
        }, 300)
      }
    }
  },
  
  // Initialize interactive elements
  initGuitarPage() {
    // Add click handlers for feature expansion
    document.querySelectorAll('[onclick*="toggleFeatureDetails"]').forEach(button => {
      button.addEventListener('click', (e) => {
        e.preventDefault()
        const featureId = button.getAttribute('onclick').match(/toggleFeatureDetails\('([^']+)'\)/)?.[1]
        if (featureId) {
          this.toggleFeatureDetails(featureId)
        }
      })
    })
  }
}

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
  console.log('Guitar Specs frontend ready with HTMX + Alpine.js + Tailwind CSS')
  
  // Initialize guitar page functionality if we're on a guitar page
  if (window.location.pathname.startsWith('/guitar/')) {
    guitarSpecs.initGuitarPage()
  }
  
  // Add any other initialization code here
  // For example, initialize tooltips, modals, etc.
})
