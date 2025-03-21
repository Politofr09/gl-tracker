#version 330

in vec2 fragTexCoord;
out vec4 fragColor;

uniform sampler2D texture0;  // Input texture
uniform float time;          // Time for animation effects
uniform vec2 resolution;     // Screen resolution

#define SCANLINE_INTENSITY 0.7
#define ABERRATION 0.001
#define CURVATURE 0.5
#define VIGNETTE_STRENGTH 0.5
#define GLITCH_INTENSITY 0  // Glitch line intensity
#define GLITCH_CHANCE 0.4      // Probability for a glitch line to appear
#define GLITCH_SPEED 10.0      // Speed of the glitch effect
#define GRAIN_INTENSITY 0.01   // Grain intensity
#define BAND_HEIGHT 0.02      // Vertical Band height
#define BAND_INTENSITY 0.2    // Band intensity

// Distorts UVs to create screen curvature (optional)
vec2 CurveUV(vec2 uv)
{
    vec2 centered = uv * 2.0 - 1.0; // Center from [-1,1]
    centered *= vec2(1.1, 1.1); // Slight zoom
    centered += centered * abs(centered) * CURVATURE * 0.01;
    return centered * 0.5 + 0.5;
}

// Simple vignette function
float Vignette(vec2 uv)
{
    float dist = distance(uv, vec2(0.5, 0.5));  
    return smoothstep(0.8, VIGNETTE_STRENGTH, dist);
}

// Adds scanlines and flicker
float Scanlines(vec2 uv, float time)
{
    float scanline = sin(uv.y * resolution.y * 2.0) * SCANLINE_INTENSITY;
    float flicker = (sin(time * 10.0) * 0.02);  
    return 1.0 - scanline + flicker;
}

// Grain effect
float Grain(vec2 uv, float time)
{
    vec2 grainSeed = vec2(12.9898, 78.233) + time * 0.1;
    return fract(sin(dot(uv * resolution.xy, grainSeed)) * 43758.5453);
}

// Adds glitch effect by shifting the fragment on the X and Y axis at random intervals
vec3 GlitchEffect(vec2 uv, float time)
{
    // Randomly shift the X-coordinate based on the Y-coordinate and time
    float glitchOffsetX = (fract(sin(time * 12.9898) * 43758.5453) - 0.5) * GLITCH_INTENSITY;
    float glitchOffsetY = (fract(cos(time * 78.233) * 43758.5453) - 0.5) * GLITCH_INTENSITY;
    uv += vec2(glitchOffsetX, glitchOffsetY);

    // RGB chromatic aberration (shifts red/blue slightly)
    float r = texture(texture0, uv + vec2(ABERRATION, 0)).r;
    float g = texture(texture0, uv).g;
    float b = texture(texture0, uv - vec2(ABERRATION, 0)).b;

    vec3 color = vec3(r, g, b);

    return color;
}

// VHS Band with distortion
vec3 VHSBandEffect(vec2 uv, float time)
{
    // Vertical banding effect
    float bandPos = fract(time * 0.2);
    if (abs(uv.y - bandPos) < BAND_HEIGHT)
    {
        // Add static noise
        float randomStatic = fract(sin(dot(uv * resolution.xy, vec2(12.9898, 78.233))) * 43758.5453);
        float bandEffect = randomStatic * BAND_INTENSITY;

        // Warp the X-axis with some choppiness
        uv.x += sin(uv.y * resolution.y * 10.0 + randomStatic) * 0.005;

        // Chromatic aberration effect (shifts red/blue slightly)
        vec3 chromaColor = vec3(
            texture(texture0, uv + vec2(0.2 * bandEffect, 0.0)).r,
            texture(texture0, uv).g,
            texture(texture0, uv - vec2(0.2 * bandEffect, 0.0)).b
        );

        // Return the chromatic band effect
        return chromaColor;
    }

    return vec3(0.0);
}

void main()
{
    vec2 uv = fragTexCoord;

    // Apply screen curvature (optional)
    // uv = CurveUV(uv);

    // Original texture color
    vec3 originalColor = texture(texture0, uv).rgb;

    // Apply glitch effect
    vec3 glitchColor = GlitchEffect(uv, time);

    // Apply VHS band effect
    vec3 bandColor = VHSBandEffect(uv, time);

    // Apply scanline and flicker effects
    float scanline = Scanlines(uv, time);

    // Apply grain effect
    float grain = Grain(uv, time);

    // Combine effects
    vec3 finalColor = mix(originalColor, glitchColor, 0.5);
    finalColor = mix(finalColor, bandColor, 0.5);
    finalColor *= scanline * 0.25;
    finalColor += grain * GRAIN_INTENSITY;
    finalColor *= vec3(5.0, 5.0, 5.0); // Increase brightness

    fragColor = vec4(finalColor, 1.0);
}
