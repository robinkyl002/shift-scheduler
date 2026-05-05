

const selected = new Set();
const minutesPerSlot = 10;

function formatMinutes(totalMinutes) {
    const hours = Math.floor(totalMinutes / 60);
    const minutes = totalMinutes % 60;
    return `${hours}h ${minutes}m`;
}

function updatedHiddenInput() {
    document.getElementById('selected-slots').value = 
    JSON.stringify(Array.from(selected));
}

function updateTotals() {
    const dayMinutes = {};
    let weeklyMinutes = 0;

    for (const key of selected) {
        const [day] = key.split("|")
        dayMinutes[day] = (dayMinutes[day] || 0) + minutesPerSlot;
        weeklyMinutes += minutesPerSlot;
    }

    document.getElementById('weekly-total').textContent = `Weekly Total: ${formatMinutes(weeklyMinutes)}`;

    document.querySelectorAll("tr[data-day]").forEach((row) => {
        const day = row.dataset.day;
        const total = dayMinutes[day] || 0
        row.querySelector(".day-total").textContent = formatMinutes(total)
    })

}

document.addEventListener("click", (event) => {
    const button = event.target.closest(".slice")
    if (!button) return;

    const key = [
        button.dataset.day,
        button.dataset.hour,
        button.dataset.slice
    ].join("|")

    if (selected.has(key)) {
        selected.delete(key)
        button.classList.remove("selected")
        button.setAttribute("aria-pressed", "false")
    } else {
        selected.add(key)
        button.classList.add("selected")
        button.setAttribute("aria")
    }

    updatedHiddenInput()
    updateTotals()
})

updatedHiddenInput()
updateTotals()