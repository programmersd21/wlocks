package ui

import (
	"time"
)

// Animation represents a smooth value transition over time
type Animation struct {
	from      float64
	to        float64
	duration  time.Duration
	startTime time.Time
	current   float64
	running   bool
	easing    EasingFunc
}

// EasingFunc defines an easing function
type EasingFunc func(t float64) float64

func EaseInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - pow(-2*t+2, 3)/2
}

func EaseOutCubic(t float64) float64 {
	return 1 - pow(1-t, 3)
}

func EaseInCubic(t float64) float64 {
	return t * t * t
}

func EaseInOutQuad(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return 1 - pow(-2*t+2, 2)/2
}

func Linear(t float64) float64 {
	return t
}

func pow(x, n float64) float64 {
	result := 1.0
	for i := 0; i < int(n); i++ {
		result *= x
	}
	return result
}

// NewAnimation creates a new animation
func NewAnimation(from, to float64, duration time.Duration, easing EasingFunc) *Animation {
	if easing == nil {
		easing = EaseInOutCubic
	}
	return &Animation{
		from:     from,
		to:       to,
		duration: duration,
		current:  from,
		easing:   easing,
	}
}

// Start begins the animation
func (a *Animation) Start() {
	a.running = true
	a.startTime = time.Now()
}

// Update updates the animation state and returns current value
func (a *Animation) Update() float64 {
	if !a.running {
		return a.current
	}

	elapsed := time.Since(a.startTime)
	if elapsed >= a.duration {
		a.current = a.to
		a.running = false
		return a.current
	}

	progress := float64(elapsed) / float64(a.duration)
	easedProgress := a.easing(progress)
	a.current = a.from + (a.to-a.from)*easedProgress
	return a.current
}

// Value returns the current value without updating
func (a *Animation) Value() float64 {
	return a.current
}

// IsRunning returns whether the animation is currently running
func (a *Animation) IsRunning() bool {
	return a.running
}

// ScrollAnimation manages smooth scrolling
type ScrollAnimation struct {
	target    int
	current   float64
	animation *Animation
}

// NewScrollAnimation creates a smooth scroll animation
func NewScrollAnimation(current int) *ScrollAnimation {
	return &ScrollAnimation{
		target:  current,
		current: float64(current),
	}
}

// SetTarget sets a new scroll target
func (s *ScrollAnimation) SetTarget(target int, duration time.Duration) {
	if target == s.target {
		return
	}
	s.target = target
	s.animation = NewAnimation(s.current, float64(target), duration, EaseOutCubic)
	s.animation.Start()
}

// Update updates the scroll position
func (s *ScrollAnimation) Update() int {
	if s.animation != nil && s.animation.IsRunning() {
		s.current = s.animation.Update()
	}
	return int(s.current + 0.5)
}

// IsAnimating returns whether scrolling is in progress
func (s *ScrollAnimation) IsAnimating() bool {
	return s.animation != nil && s.animation.IsRunning()
}

// FadeAnimation manages opacity transitions
type FadeAnimation struct {
	visible   bool
	opacity   float64
	animation *Animation
}

// NewFadeAnimation creates a fade animation
func NewFadeAnimation() *FadeAnimation {
	return &FadeAnimation{
		visible: false,
		opacity: 0.0,
	}
}

// FadeIn starts fading in
func (f *FadeAnimation) FadeIn(duration time.Duration) {
	f.visible = true
	f.animation = NewAnimation(f.opacity, 1.0, duration, EaseInOutQuad)
	f.animation.Start()
}

// FadeOut starts fading out
func (f *FadeAnimation) FadeOut(duration time.Duration) {
	f.animation = NewAnimation(f.opacity, 0.0, duration, EaseInOutQuad)
	f.animation.Start()
}

// Update updates the fade state
func (f *FadeAnimation) Update() float64 {
	if f.animation != nil && f.animation.IsRunning() {
		f.opacity = f.animation.Update()
		if f.opacity <= 0.0 {
			f.visible = false
		}
	}
	return f.opacity
}

// Opacity returns the current opacity (0.0 to 1.0)
func (f *FadeAnimation) Opacity() float64 {
	return f.opacity
}

// IsVisible returns whether the element should be visible
func (f *FadeAnimation) IsVisible() bool {
	return f.visible || f.opacity > 0.0
}
