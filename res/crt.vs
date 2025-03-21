#version 330

in vec3 vertexPosition;
in vec2 vertexTexCoord;
out vec2 fragTexCoord;

uniform mat4 mvp;
uniform float time;

void main() {
    gl_Position = mvp * vec4(vertexPosition, 1.0); // Transform the vertex position
    
    fragTexCoord = vertexTexCoord;
}
