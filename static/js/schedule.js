// const table = document.getElementById("schedule-table")
const selected = new Set();
const minutesPerSlot = 10;
let isDragging = false
let dragMode = null

function formatMinutes(totalMinutes) {
    const hours = Math.floor(totalMinutes / 60);
    const minutes = totalMinutes % 60;
    return `${hours}h ${minutes}m`;
}

function updateHiddenInput() {
    const input = document.getElementById('selected-slots')
    if (!input) return;
    input.value = JSON.stringify(Array.from(selected));
}

function updateTotals() {
    const weeklyTotal = document.getElementById('weekly-total')
    if(!weeklyTotal) return;

    const dayMinutes = {};
    let weeklyMinutes = 0;

    for (const key of selected) {
        const [day] = key.split("|")
        dayMinutes[day] = (dayMinutes[day] || 0) + minutesPerSlot;
        weeklyMinutes += minutesPerSlot;
    }

    weeklyTotal.textContent = `Weekly Total: ${formatMinutes(weeklyMinutes)}`;

    document.querySelectorAll("tr[data-day]").forEach((row) => {
        const day = row.dataset.day;
        const total = dayMinutes[day] || 0
        row.querySelector(".day-total").textContent = formatMinutes(total)
    })

}

function syncSelectedFromDOM() {
    selected.clear()

    document.querySelectorAll(".slice.selected").forEach((button) => {
        selected.add(getSlotKey(button))
        button.setAttribute("aria-pressed", "true")
    })

    document.querySelectorAll(".slice:not(.selected)").forEach((button) => {
        button.setAttribute("aria-pressed", "false")
    })
}

function initializeScheduleUI() {
    syncSelectedFromDOM()
    updateHiddenInput()
    updateTotals()
}

function getSlotKey(button) {
    return [
        button.dataset.day,
        button.dataset.hour,
        button.dataset.slice
    ].join("|")
}

function setSliceState(button, shouldSelect) {
    const key = getSlotKey(button)
    
    if (shouldSelect && selected.has(key)) return;
    if (!shouldSelect && !selected.has(key)) return;

    if (!shouldSelect) {
        selected.delete(key)
        button.classList.remove("selected")
        button.setAttribute("aria-pressed", "false")
    } else {
        selected.add(key)
        button.classList.add("selected")
        button.setAttribute("aria-pressed", "true")
    }

    updateHiddenInput()
    updateTotals()
}

document.addEventListener("pointerdown", (event) => {
    const button = event.target.closest(".slice")
    if (!button) return;

    event.preventDefault()

    const key = getSlotKey(button)
    isDragging = true
    dragMode = selected.has(key) ? "deselect" : "select"

    setSliceState(button, dragMode === 'select')

})

document.addEventListener("pointerover", (event) => {
    if (!isDragging) return;

    const button = event.target.closest(".slice")
    if (!button) return;

    setSliceState(button, dragMode === 'select')
})

document.addEventListener('pointerup', () => {
    isDragging = false
    dragMode = null
})

document.addEventListener("DOMContentLoaded", initializeScheduleUI)
document.body.addEventListener("htmx:afterSwap", initializeScheduleUI)
