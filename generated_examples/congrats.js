/**
 * ROT13 decoder script
 * Decodes the given ROT13-encoded string and prints the result
 */

// Function to decode a ROT13 encoded string
function rot13Decode(encodedStr) {
  return encodedStr.replace(/[a-zA-Z]/g, function(char) {
    // Get the ASCII code
    const code = char.charCodeAt(0);
    
    // Handle uppercase letters (ASCII 65-90)
    if (code >= 65 && code <= 90) {
      return String.fromCharCode(((code - 65 + 13) % 26) + 65);
    }
    
    // Handle lowercase letters (ASCII 97-122)
    if (code >= 97 && code <= 122) {
      return String.fromCharCode(((code - 97 + 13) % 26) + 97);
    }
    
    // Return non-alphabetic characters unchanged
    return char;
  });
}

// The ROT13-encoded string
const encodedString = 'Pbatenghyngvbaf ba ohvyqvat n pbqr-rqvgvat ntrag!';

// Decode and print the message
const decodedString = rot13Decode(encodedString);
console.log(decodedString);